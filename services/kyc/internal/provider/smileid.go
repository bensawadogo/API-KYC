package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/datakeys/kyc-service/config"
	"github.com/datakeys/kyc-service/internal"
	"github.com/datakeys/kyc-service/internal/model"
)

type SmileIDProvider struct {
	cfg      config.SmileIDConfig
	client   *http.Client
	testBaseURL string // overrides baseURL() when set (testing only)
}

func NewSmileIDProvider(cfg config.SmileIDConfig) *SmileIDProvider {
	return &SmileIDProvider{
		cfg: cfg,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (p *SmileIDProvider) Name() string {
	return "smileid"
}

func (p *SmileIDProvider) SupportedCountries() []string {
	return []string{
		"BF", "SN", "ML", "CI", "GH", "NG", "KE", "TZ", "UG", "CM", "ZA", "ZM", "ZW",
		"BJ", "TG", "GN", "SL", "NE", "GW", "LR", "GM", "MR", "CV",
	}
}

func (p *SmileIDProvider) baseURL() string {
	if p.testBaseURL != "" {
		return p.testBaseURL
	}
	if p.cfg.Sandbox {
		return "https://testapi.smileidentity.com/v1/"
	}
	return "https://api.smileidentity.com/v1/"
}

func (p *SmileIDProvider) Verify(ctx context.Context, req internal.ProviderRequest) (*internal.ProviderResult, error) {
	jobType := 6
	if len(req.DocData) > 0 && isSelfieData(req.DocData) {
		jobType = 2
	}

	payload := map[string]interface{}{
		"partner_id":        p.cfg.PartnerId,
		"source_sdk":        "rest_api",
		"source_sdk_version": "1.0.0",
		"job_type":          jobType,
		"user_id":           req.VerificationID,
		"country":           req.CountryCode,
		"id_type":           mapDocType(req.DocType),
		"id_number":         req.DocNumber,
		"callback_url":      p.cfg.CallbackUrl,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("smileid marshal payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL()+"id_verification", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("smileid create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.cfg.ApiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("smileid request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("smileid read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("smileid API error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var apiResp smileIDResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("smileid parse response: %w", err)
	}

	score := apiResp.ConfidenceValue / 100.0
	flags := parseSmileIDActions(apiResp.Actions)
	approved := score >= 0.70

	if score < 0.70 {
		flags = append(flags, model.FlagLowConfidence)
		approved = false
	}

	rawData := map[string]interface{}{}
	_ = json.Unmarshal(respBody, &rawData)

	return &internal.ProviderResult{
		Approved: approved,
		Score:    score,
		Flags:    flags,
		Provider: p.Name(),
		RawData:  rawData,
	}, nil
}

type smileIDResponse struct {
	ConfidenceValue float64                `json:"confidence_value"`
	Actions         map[string]interface{} `json:"Actions"`
	Result          string                 `json:"Result"`
}

func parseSmileIDActions(actions map[string]interface{}) []string {
	if actions == nil {
		return nil
	}

	flags := make([]string, 0)
	for key, val := range actions {
		strVal, ok := val.(string)
		if !ok {
			continue
		}
		if strings.EqualFold(strVal, "failed") || strings.EqualFold(strVal, "rejected") {
			switch strings.ToLower(key) {
			case "expired":
				flags = append(flags, model.FlagExpiredDoc)
			case "sanctions":
				flags = append(flags, model.FlagSanctionsMatch)
			case "format", "id_format":
				flags = append(flags, model.FlagInvalidFormat)
			default:
				flags = append(flags, model.FlagLowConfidence)
			}
		}
	}
	return flags
}

func mapDocType(docType string) string {
	switch strings.ToUpper(docType) {
	case "NATIONAL_ID":
		return "NATIONAL_ID"
	case "PASSPORT":
		return "PASSPORT"
	case "DRIVERS_LICENSE":
		return "DRIVERS_LICENSE"
	case "VOTER_CARD":
		return "VOTER_ID"
	default:
		return strings.ToUpper(docType)
	}
}

func isSelfieData(data []byte) bool {
	var meta map[string]interface{}
	if err := json.Unmarshal(data, &meta); err != nil {
		return false
	}
	_, hasSelfie := meta["selfie"]
	return hasSelfie
}
