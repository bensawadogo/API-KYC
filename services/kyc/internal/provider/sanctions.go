package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/datakeys/kyc-service/internal"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type OpenSanctionsProvider struct {
	baseURL   string
	apiKey    string
	threshold float64
	client    *http.Client
	logger    *zap.Logger
}

func NewOpenSanctionsProvider(apiKey, baseURL string, threshold float64, logger *zap.Logger) *OpenSanctionsProvider {
	if threshold <= 0 {
		threshold = 0.70
	}
	if baseURL == "" {
		baseURL = "https://api.opensanctions.org"
	}
	return &OpenSanctionsProvider{
		baseURL:   baseURL,
		apiKey:    apiKey,
		threshold: threshold,
		client:    &http.Client{Timeout: 10 * time.Second},
		logger:    logger,
	}
}

func (p *OpenSanctionsProvider) Name() string {
	return "opensanctions"
}

func (p *OpenSanctionsProvider) Check(ctx context.Context, req internal.AMLRequest) (*internal.AMLResult, error) {
	qID := "kyc-" + uuid.New().String()

	props := map[string]interface{}{
		"name":    []string{req.FullName},
		"country": []string{req.CountryCode},
	}
	if req.DateOfBirth != "" {
		props["birthDate"] = []string{req.DateOfBirth}
	}

	body := map[string]interface{}{
		"queries": map[string]interface{}{
			qID: map[string]interface{}{
				"schema":     "Person",
				"properties": props,
			},
		},
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("opensanctions marshal: %w", err)
	}

	url := p.baseURL + "/match/default"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("opensanctions create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "ApiKey "+p.apiKey)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("opensanctions request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("opensanctions API error: status=%d", resp.StatusCode)
	}

	var apiResp struct {
		Responses map[string]struct {
			Results []struct {
				Score   float64  `json:"score"`
				Match   bool     `json:"match"`
				Topics  []string `json:"topics"`
				Caption string   `json:"caption"`
				ID      string   `json:"id"`
				Dataset struct {
					Name string `json:"name"`
				} `json:"dataset"`
			} `json:"results"`
		} `json:"responses"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("opensanctions decode: %w", err)
	}

	response, ok := apiResp.Responses[qID]
	if !ok {
		p.logger.Warn("opensanctions: missing response for query",
			zap.String("query_id", qID))
		return &internal.AMLResult{
			Score:        0,
			IsSanctioned: false,
			IsPEP:        false,
			ScreenedAt:   time.Now().UTC(),
			Source:       p.Name(),
		}, nil
	}

	result := &internal.AMLResult{
		Matches:    make([]internal.AMLMatch, 0),
		ScreenedAt: time.Now().UTC(),
		Source:     p.Name(),
	}

	maxScore := 0.0
	for _, r := range response.Results {
		if r.Score >= p.threshold {
			entry := internal.AMLMatch{
				EntityName: r.Caption,
				EntityID:   r.ID,
				Topics:     r.Topics,
				Score:      r.Score,
				Dataset:    r.Dataset.Name,
			}
			result.Matches = append(result.Matches, entry)
			if r.Score > maxScore {
				maxScore = r.Score
			}
			for _, topic := range r.Topics {
				switch topic {
				case "sanction":
					result.IsSanctioned = true
				case "pep":
					result.IsPEP = true
				}
			}
		}
	}

	result.Score = maxScore

	p.logger.Info("aml screening completed",
		zap.Int("matches", len(result.Matches)),
		zap.Float64("score", result.Score),
	)

	return result, nil
}

type LocalAMLProvider struct{}

func NewLocalAMLProvider() *LocalAMLProvider {
	return &LocalAMLProvider{}
}

func (p *LocalAMLProvider) Name() string {
	return "local_aml"
}

func (p *LocalAMLProvider) Check(_ context.Context, _ internal.AMLRequest) (*internal.AMLResult, error) {
	return &internal.AMLResult{
		IsSanctioned: false,
		IsPEP:        false,
		Score:        0,
		ScreenedAt:   time.Now().UTC(),
		Source:       p.Name(),
	}, nil
}
