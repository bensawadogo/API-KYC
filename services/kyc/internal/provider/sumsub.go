package provider

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/datakeys/kyc-service/config"
	"github.com/datakeys/kyc-service/internal"
	"github.com/datakeys/kyc-service/internal/model"
)

type SumSubProvider struct {
	cfg    config.SumSubConfig
	client *http.Client
}

func NewSumSubProvider(cfg config.SumSubConfig) *SumSubProvider {
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.sumsub.com"
	}
	return &SumSubProvider{
		cfg: cfg,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (p *SumSubProvider) Name() string {
	return "sumsub"
}

func (p *SumSubProvider) SupportedCountries() []string {
	return []string{"MA", "EG", "TN", "DZ", "LY", "SD"}
}

func (p *SumSubProvider) Verify(ctx context.Context, req internal.ProviderRequest) (*internal.ProviderResult, error) {
	applicantID, err := p.createApplicant(ctx, req)
	if err != nil {
		return nil, err
	}

	if len(req.DocData) > 0 {
		if err := p.uploadDocument(ctx, applicantID, req); err != nil {
			return nil, err
		}
	}

	if err := p.requestReview(ctx, applicantID); err != nil {
		return nil, err
	}

	reviewAnswer, rawData, err := p.pollReviewStatus(ctx, applicantID)
	if err != nil {
		return nil, err
	}

	result := mapSumSubReview(reviewAnswer)
	result.Provider = p.Name()
	result.RawData = rawData
	return result, nil
}

func (p *SumSubProvider) createApplicant(ctx context.Context, req internal.ProviderRequest) (string, error) {
	payload := map[string]interface{}{
		"externalUserId": req.VerificationID,
		"info": map[string]string{
			"country": strings.ToUpper(req.CountryCode),
			"phone":   req.Phone,
		},
	}

	body, _ := json.Marshal(payload)
	path := "/resources/applicants?levelName=basic-kyc-level"
	respBody, err := p.doRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return "", fmt.Errorf("sumsub create applicant: %w", err)
	}

	var resp struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(respBody, &resp); err != nil {
		return "", fmt.Errorf("sumsub parse applicant: %w", err)
	}
	if resp.ID == "" {
		return "", fmt.Errorf("sumsub applicant id empty")
	}
	return resp.ID, nil
}

func (p *SumSubProvider) uploadDocument(ctx context.Context, applicantID string, req internal.ProviderRequest) error {
	meta := map[string]string{
		"idDocType":  mapSumSubDocType(req.DocType),
		"country":    strings.ToUpper(req.CountryCode),
		"number":     req.DocNumber,
	}
	metaBody, _ := json.Marshal(meta)

	path := fmt.Sprintf("/resources/applicants/%s/info/idDoc", applicantID)
	body := append(metaBody, '\n')
	body = append(body, req.DocData...)

	_, err := p.doRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return fmt.Errorf("sumsub upload document: %w", err)
	}
	return nil
}

func (p *SumSubProvider) requestReview(ctx context.Context, applicantID string) error {
	path := fmt.Sprintf("/resources/applicants/%s/status/pending", applicantID)
	_, err := p.doRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return fmt.Errorf("sumsub request review: %w", err)
	}
	return nil
}

func (p *SumSubProvider) pollReviewStatus(ctx context.Context, applicantID string) (string, map[string]interface{}, error) {
	path := fmt.Sprintf("/resources/applicants/%s/requiredIdDocsStatus", applicantID)

	for attempt := 0; attempt < 10; attempt++ {
		respBody, err := p.doRequest(ctx, http.MethodGet, path, nil)
		if err != nil {
			return "", nil, err
		}

		rawData := map[string]interface{}{}
		_ = json.Unmarshal(respBody, &rawData)

		var review struct {
			Review struct {
				ReviewResult struct {
					ReviewAnswer string `json:"reviewAnswer"`
				} `json:"reviewResult"`
			} `json:"review"`
		}
		if err := json.Unmarshal(respBody, &review); err == nil {
			answer := review.Review.ReviewResult.ReviewAnswer
			if answer != "" {
				return answer, rawData, nil
			}
		}

		select {
		case <-ctx.Done():
			return "", nil, ctx.Err()
		case <-time.After(3 * time.Second):
		}
	}

	return "RETRY", map[string]interface{}{"status": "timeout"}, nil
}

func mapSumSubReview(answer string) *internal.ProviderResult {
	switch strings.ToUpper(answer) {
	case "GREEN":
		return &internal.ProviderResult{
			Approved: true,
			Score:    0.92,
			Flags:    nil,
		}
	case "RED":
		return &internal.ProviderResult{
			Approved: false,
			Score:    0.15,
			Flags:    []string{model.FlagLowConfidence},
		}
	default:
		return &internal.ProviderResult{
			Approved: false,
			Score:    0.50,
			Flags:    []string{model.FlagManualRequired},
		}
	}
}

func mapSumSubDocType(docType string) string {
	switch strings.ToUpper(docType) {
	case "NATIONAL_ID":
		return "ID_CARD"
	case "PASSPORT":
		return "PASSPORT"
	case "DRIVERS_LICENSE":
		return "DRIVERS"
	case "RESIDENCE_PERMIT":
		return "RESIDENCE_PERMIT"
	default:
		return strings.ToUpper(docType)
	}
}

func (p *SumSubProvider) doRequest(ctx context.Context, method, path string, body []byte) ([]byte, error) {
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	sig := p.sign(method, path, ts, body)

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}

	url := p.cfg.BaseURL + path
	httpReq, err := http.NewRequestWithContext(ctx, method, url, reader)
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("X-App-Token", p.cfg.AppToken)
	httpReq.Header.Set("X-App-Access-Ts", ts)
	httpReq.Header.Set("X-App-Access-Sig", sig)
	if body != nil {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("sumsub API error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

func (p *SumSubProvider) sign(method, path, ts string, body []byte) string {
	payload := ts + method + path
	if body != nil {
		payload += string(body)
	}
	mac := hmac.New(sha256.New, []byte(p.cfg.SecretKey))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}
