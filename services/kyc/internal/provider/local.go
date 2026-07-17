package provider

import (
	"context"
	"strings"

	"github.com/datakeys/kyc-service/internal"
	"github.com/datakeys/kyc-service/internal/countries"
	"github.com/datakeys/kyc-service/internal/model"
)

type LocalProvider struct {
	sandbox bool
}

func NewLocalProvider() *LocalProvider {
	return &LocalProvider{sandbox: true}
}

func NewLocalProviderWithSandbox(sandbox bool) *LocalProvider {
	return &LocalProvider{sandbox: sandbox}
}

func (p *LocalProvider) Name() string {
	return "local"
}

func (p *LocalProvider) SupportedCountries() []string {
	return countries.AllAfricanCountryCodes()
}

func (p *LocalProvider) Verify(ctx context.Context, req internal.ProviderRequest) (*internal.ProviderResult, error) {
	_ = ctx

	if p.sandbox {
		return &internal.ProviderResult{
			Approved: true,
			Score:    0.99,
			Flags:    nil,
			Provider: p.Name(),
			RawData:  map[string]interface{}{"mode": "sandbox"},
		}, nil
	}

	if req.DocNumber != "" && !countries.ValidateDocNumber(req.CountryCode, req.DocType, req.DocNumber) {
		return &internal.ProviderResult{
			Approved: false,
			Score:    0.2,
			Flags:    []string{model.FlagInvalidFormat},
			Provider: p.Name(),
		}, nil
	}

	return &internal.ProviderResult{
		Approved: true,
		Score:    0.95,
		Flags:    nil,
		Provider: p.Name(),
		RawData:  map[string]interface{}{"mode": "local"},
	}, nil
}

// IsFallbackCandidate returns true when the provider can handle any country.
func (p *LocalProvider) IsFallbackCandidate(countryCode string) bool {
	code := strings.ToUpper(countryCode)
	for _, c := range p.SupportedCountries() {
		if c == code {
			return true
		}
	}
	return false
}
