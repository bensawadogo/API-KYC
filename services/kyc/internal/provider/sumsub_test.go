package provider_test

import (
	"testing"

	"github.com/datakeys/kyc-service/config"
	"github.com/datakeys/kyc-service/internal/provider"
)

func TestSumSubProvider_Name(t *testing.T) {
	p := provider.NewSumSubProvider(config.SumSubConfig{})
	if p.Name() != "sumsub" {
		t.Errorf("expected sumsub, got %s", p.Name())
	}
}

func TestSumSubProvider_SupportedCountries(t *testing.T) {
	p := provider.NewSumSubProvider(config.SumSubConfig{})
	countries := p.SupportedCountries()
	if len(countries) == 0 {
		t.Fatal("expected at least one country")
	}
	hasMA := false
	for _, c := range countries {
		if c == "MA" {
			hasMA = true
			break
		}
	}
	if !hasMA {
		t.Error("expected MA in supported countries")
	}
}

func TestSumSubProvider_DefaultBaseURL(t *testing.T) {
	p := provider.NewSumSubProvider(config.SumSubConfig{
		AppToken:  "test-token",
		SecretKey: "test-secret",
	})
	if p.Name() != "sumsub" {
		t.Errorf("expected sumsub, got %s", p.Name())
	}
}

func TestSumSubProvider_CustomBaseURL(t *testing.T) {
	p := provider.NewSumSubProvider(config.SumSubConfig{
		AppToken:  "test-token",
		SecretKey: "test-secret",
		BaseURL:   "https://custom.sumsub.com",
	})
	if p.Name() != "sumsub" {
		t.Errorf("expected sumsub, got %s", p.Name())
	}
}
