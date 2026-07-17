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

type YouverifyProvider struct {
	cfg    config.YouverifyConfig
	client *http.Client
}

func NewYouverifyProvider(cfg config.YouverifyConfig) *YouverifyProvider {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		if cfg.Sandbox {
			baseURL = "https://api.staging.youverify.co/v2/"
		} else {
			baseURL = "https://api.youverify.co/v2/"
		}
	}
	cfg.BaseURL = baseURL

	return &YouverifyProvider{
		cfg: cfg,
		client: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

func (p *YouverifyProvider) Name() string {
	return "youverify"
}

func (p *YouverifyProvider) SupportedCountries() []string {
	return []string{"NG", "GH"}
}

func (p *YouverifyProvider) Verify(ctx context.Context, req internal.ProviderRequest) (*internal.ProviderResult, error) {
	endpoint, payload := p.resolveEndpoint(req)
	if endpoint == "" {
		return &internal.ProviderResult{
			Approved: false,
			Score:    0.1,
			Flags:    []string{model.FlagUnsupportedDoc},
			Provider: p.Name(),
		}, nil
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("youverify marshal payload: %w", err)
	}

	url := p.cfg.BaseURL + endpoint
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("youverify create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("token", p.cfg.ApiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("youverify request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("youverify read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("youverify API error: status=%d body=%s", resp.StatusCode, string(respBody))
	}

	var apiResp youverifyResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("youverify parse response: %w", err)
	}

	rawData := map[string]interface{}{}
	_ = json.Unmarshal(respBody, &rawData)

	if strings.EqualFold(apiResp.Data.MatchStatus, "found") {
		return &internal.ProviderResult{
			Approved: true,
			Score:    0.95,
			Flags:    nil,
			Provider: p.Name(),
			RawData:  rawData,
		}, nil
	}

	return &internal.ProviderResult{
		Approved: false,
		Score:    0.1,
		Flags:    []string{model.FlagLowConfidence},
		Provider: p.Name(),
		RawData:  rawData,
	}, nil
}

type youverifyResponse struct {
	Success bool `json:"success"`
	Data    struct {
		MatchStatus string `json:"matchStatus"`
	} `json:"data"`
}

func (p *YouverifyProvider) resolveEndpoint(req internal.ProviderRequest) (string, map[string]string) {
	country := strings.ToUpper(req.CountryCode)
	docType := strings.ToUpper(req.DocType)

	switch country {
	case "NG":
		switch docType {
		case "NATIONAL_ID":
			return "identities/nin", map[string]string{"id": req.DocNumber}
		case "BVN":
			return "identities/bvn", map[string]string{"id": req.DocNumber}
		}
	case "GH":
		if docType == "NATIONAL_ID" {
			return "identities/gh/ghana-card", map[string]string{"id": req.DocNumber}
		}
	}

	return "", nil
}
