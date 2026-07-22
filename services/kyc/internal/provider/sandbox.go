package provider

import (
	"context"
	"math/rand"
	"strings"
	"time"

	"github.com/datakeys/kyc-service/internal"
	"github.com/datakeys/kyc-service/internal/countries"
	"github.com/datakeys/kyc-service/internal/model"
)

type SandboxProvider struct{}

func NewSandboxProvider() *SandboxProvider {
	return &SandboxProvider{}
}

func (p *SandboxProvider) Name() string {
	return "sandbox"
}

func (p *SandboxProvider) SupportedCountries() []string {
	return countries.AllAfricanCountryCodes()
}

func (p *SandboxProvider) Verify(ctx context.Context, req internal.ProviderRequest) (*internal.ProviderResult, error) {
	delay := 800*time.Millisecond + time.Duration(rand.Intn(200))*time.Millisecond
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(delay):
	}

	docNum := strings.TrimSpace(req.DocNumber)

	switch {
	case strings.HasSuffix(docNum, "0000"):
		return approved("sandbox", 0.95), nil

	case strings.HasSuffix(docNum, "1111"):
		return &internal.ProviderResult{
			Approved: false,
			Score:    0.15,
			Flags:    []string{model.FlagLowConfidence},
			Provider: p.Name(),
			RawData:  map[string]interface{}{"mode": "sandbox", "scenario": "rejected"},
		}, nil

	case strings.HasSuffix(docNum, "2222"):
		return &internal.ProviderResult{
			Approved: false,
			Score:    0.05,
			Flags:    []string{model.FlagSanctionsMatch},
			Provider: p.Name(),
			RawData:  map[string]interface{}{"mode": "sandbox", "scenario": "sanctions"},
		}, nil

	case strings.HasSuffix(docNum, "3333"):
		return &internal.ProviderResult{
			Approved: false,
			Score:    0.45,
			Flags:    []string{model.FlagPEPDetected},
			Provider: p.Name(),
			RawData:  map[string]interface{}{"mode": "sandbox", "scenario": "pep"},
		}, nil

	case strings.HasSuffix(docNum, "4444"):
		return &internal.ProviderResult{
			Approved: false,
			Score:    0.10,
			Flags:    []string{model.FlagExpiredDoc, model.FlagInvalidFormat},
			Provider: p.Name(),
			RawData:  map[string]interface{}{"mode": "sandbox", "scenario": "expired_doc"},
		}, nil

	default:
		return approved("sandbox", 0.92), nil
	}
}

func approved(provider string, score float64) *internal.ProviderResult {
	return &internal.ProviderResult{
		Approved: true,
		Score:    score,
		Flags:    nil,
		Provider: provider,
		RawData:  map[string]interface{}{"mode": "sandbox"},
	}
}
