package provider_test

import (
	"testing"

	"github.com/datakeys/kyc-service/config"
	"github.com/datakeys/kyc-service/internal/provider"
)

func TestSmileIDProvider_Name(t *testing.T) {
	p := provider.NewSmileIDProvider(config.SmileIDConfig{})
	if p.Name() != "smileid" {
		t.Errorf("expected smileid, got %s", p.Name())
	}
}

func TestSmileIDProvider_SupportedCountries(t *testing.T) {
	p := provider.NewSmileIDProvider(config.SmileIDConfig{})
	countries := p.SupportedCountries()
	if len(countries) == 0 {
		t.Fatal("expected at least one country")
	}
	hasBF := false
	for _, c := range countries {
		if c == "BF" {
			hasBF = true
			break
		}
	}
	if !hasBF {
		t.Error("expected BF in supported countries")
	}
}

func TestSmileIDProvider_SandboxBaseURL(t *testing.T) {
	p := provider.NewSmileIDProvider(config.SmileIDConfig{
		ApiKey:    "test-key",
		PartnerId: "test-partner",
		Sandbox:   true,
	})
	if p.Name() != "smileid" {
		t.Errorf("expected smileid, got %s", p.Name())
	}
}

func TestSmileIDProvider_ProductionBaseURL(t *testing.T) {
	p := provider.NewSmileIDProvider(config.SmileIDConfig{
		ApiKey:    "prod-key",
		PartnerId: "prod-partner",
		Sandbox:   false,
	})
	if p.Name() != "smileid" {
		t.Errorf("expected smileid, got %s", p.Name())
	}
}
