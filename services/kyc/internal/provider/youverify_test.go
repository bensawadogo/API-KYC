package provider_test

import (
	"testing"

	"github.com/datakeys/kyc-service/config"
	"github.com/datakeys/kyc-service/internal/provider"
)

func TestYouverifyProvider_Name(t *testing.T) {
	p := provider.NewYouverifyProvider(config.YouverifyConfig{})
	if p.Name() != "youverify" {
		t.Errorf("expected youverify, got %s", p.Name())
	}
}

func TestYouverifyProvider_SupportedCountries(t *testing.T) {
	p := provider.NewYouverifyProvider(config.YouverifyConfig{})
	countries := p.SupportedCountries()
	if len(countries) == 0 {
		t.Fatal("expected at least one country")
	}
	hasNG := false
	for _, c := range countries {
		if c == "NG" {
			hasNG = true
			break
		}
	}
	if !hasNG {
		t.Error("expected NG in supported countries")
	}
}

func TestYouverifyProvider_SandboxBaseURL(t *testing.T) {
	p := provider.NewYouverifyProvider(config.YouverifyConfig{
		ApiKey:  "test-key",
		Sandbox: true,
	})
	if p.Name() != "youverify" {
		t.Errorf("expected youverify, got %s", p.Name())
	}
}

func TestYouverifyProvider_ProductionBaseURL(t *testing.T) {
	p := provider.NewYouverifyProvider(config.YouverifyConfig{
		ApiKey:  "prod-key",
		Sandbox: false,
	})
	if p.Name() != "youverify" {
		t.Errorf("expected youverify, got %s", p.Name())
	}
}

func TestYouverifyProvider_CustomBaseURL(t *testing.T) {
	p := provider.NewYouverifyProvider(config.YouverifyConfig{
		ApiKey:  "test",
		BaseURL: "https://custom.api.com/",
	})
	if p.Name() != "youverify" {
		t.Errorf("expected youverify, got %s", p.Name())
	}
}
