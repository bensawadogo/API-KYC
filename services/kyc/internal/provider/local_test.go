package provider_test

import (
	"context"
	"testing"

	"github.com/datakeys/kyc-service/internal"
	"github.com/datakeys/kyc-service/internal/provider"
)

func TestLocalProvider_Name(t *testing.T) {
	p := provider.NewLocalProvider()
	if p.Name() != "local" {
		t.Errorf("expected 'local', got '%s'", p.Name())
	}
}

func TestLocalProvider_SupportedCountries(t *testing.T) {
	p := provider.NewLocalProvider()
	countries := p.SupportedCountries()
	if len(countries) < 20 {
		t.Errorf("expected at least 20 countries, got %d", len(countries))
	}
	expected := map[string]bool{"BF": false, "NG": false, "MA": false}
	for _, c := range countries {
		if _, ok := expected[c]; ok {
			expected[c] = true
		}
	}
	for code, found := range expected {
		if !found {
			t.Errorf("expected country %s not found", code)
		}
	}
}

func TestLocalProvider_Verify_ValidCNIB(t *testing.T) {
	p := provider.NewLocalProviderWithSandbox(false)
	req := internal.ProviderRequest{
		CountryCode: "BF",
		DocType:     "NATIONAL_ID",
		DocNumber:   "B1234567",
	}
	result, err := p.Verify(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Approved {
		t.Error("expected approved for valid CNIB")
	}
	if result.Score < 0.9 {
		t.Errorf("expected score >= 0.9, got %f", result.Score)
	}
	if result.Provider != "local" {
		t.Errorf("expected provider 'local', got '%s'", result.Provider)
	}
}

func TestLocalProvider_Verify_InvalidFormat(t *testing.T) {
	p := provider.NewLocalProviderWithSandbox(false)
	req := internal.ProviderRequest{
		CountryCode: "BF",
		DocType:     "NATIONAL_ID",
		DocNumber:   "INVALID123",
	}
	result, err := p.Verify(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if result.Approved {
		t.Error("expected not approved for invalid format")
	}
	found := false
	for _, f := range result.Flags {
		if f == "INVALID_FORMAT" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected INVALID_FORMAT flag")
	}
}

func TestLocalProvider_Verify_ValidNIN_Nigeria(t *testing.T) {
	p := provider.NewLocalProviderWithSandbox(false)
	req := internal.ProviderRequest{
		CountryCode: "NG",
		DocType:     "NATIONAL_ID",
		DocNumber:   "12345678901",
	}
	result, err := p.Verify(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Approved {
		t.Error("expected approved for valid NIN")
	}
}

func TestLocalProvider_Verify_ValidPassport(t *testing.T) {
	p := provider.NewLocalProviderWithSandbox(false)
	req := internal.ProviderRequest{
		CountryCode: "BF",
		DocType:     "PASSPORT",
		DocNumber:   "AB123456",
	}
	result, err := p.Verify(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Approved {
		t.Error("expected approved for passport")
	}
}

func TestLocalProvider_Verify_EmptyDocNumber(t *testing.T) {
	p := provider.NewLocalProviderWithSandbox(true)
	req := internal.ProviderRequest{
		CountryCode: "BF",
		DocType:     "NATIONAL_ID",
		DocNumber:   "",
	}
	result, err := p.Verify(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Approved {
		t.Error("expected approved in sandbox mode")
	}
	if result.Score < 0.9 {
		t.Errorf("expected score >= 0.9, got %f", result.Score)
	}
}

func TestLocalProvider_Verify_UnknownCountry(t *testing.T) {
	p := provider.NewLocalProviderWithSandbox(true)
	req := internal.ProviderRequest{
		CountryCode: "XX",
		DocType:     "PASSPORT",
	}
	result, err := p.Verify(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Approved {
		t.Error("expected approved for unknown country in sandbox")
	}
	if result.Score < 0.9 {
		t.Errorf("expected score >= 0.9, got %f", result.Score)
	}
}