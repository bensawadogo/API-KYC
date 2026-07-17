package provider

import (
	"testing"

	"github.com/datakeys/kyc-service/config"
	"github.com/datakeys/kyc-service/internal"
)

func TestYouverifyProvider_ResolveEndpoint_NG_NationalID(t *testing.T) {
	p := NewYouverifyProvider(config.YouverifyConfig{})
	endpoint, payload := p.resolveEndpoint(internal.ProviderRequest{
		CountryCode: "NG",
		DocType:     "NATIONAL_ID",
		DocNumber:   "12345678901",
	})
	if endpoint != "identities/nin" {
		t.Errorf("expected identities/nin, got %s", endpoint)
	}
	if payload["id"] != "12345678901" {
		t.Errorf("expected id 12345678901, got %s", payload["id"])
	}
}

func TestYouverifyProvider_ResolveEndpoint_NG_BVN(t *testing.T) {
	p := NewYouverifyProvider(config.YouverifyConfig{})
	endpoint, payload := p.resolveEndpoint(internal.ProviderRequest{
		CountryCode: "NG",
		DocType:     "BVN",
		DocNumber:   "22345678901",
	})
	if endpoint != "identities/bvn" {
		t.Errorf("expected identities/bvn, got %s", endpoint)
	}
	if payload["id"] != "22345678901" {
		t.Errorf("expected id 22345678901, got %s", payload["id"])
	}
}

func TestYouverifyProvider_ResolveEndpoint_GH_NationalID(t *testing.T) {
	p := NewYouverifyProvider(config.YouverifyConfig{})
	endpoint, payload := p.resolveEndpoint(internal.ProviderRequest{
		CountryCode: "GH",
		DocType:     "NATIONAL_ID",
		DocNumber:   "GHA-123456",
	})
	if endpoint != "identities/gh/ghana-card" {
		t.Errorf("expected identities/gh/ghana-card, got %s", endpoint)
	}
	if payload["id"] != "GHA-123456" {
		t.Errorf("expected id GHA-123456, got %s", payload["id"])
	}
}

func TestYouverifyProvider_ResolveEndpoint_Unsupported(t *testing.T) {
	p := NewYouverifyProvider(config.YouverifyConfig{})
	endpoint, payload := p.resolveEndpoint(internal.ProviderRequest{
		CountryCode: "BF",
		DocType:     "NATIONAL_ID",
	})
	if endpoint != "" {
		t.Errorf("expected empty endpoint, got %s", endpoint)
	}
	if payload != nil {
		t.Errorf("expected nil payload, got %v", payload)
	}
}

func TestYouverifyProvider_ResolveEndpoint_GH_UnsupportedDoc(t *testing.T) {
	p := NewYouverifyProvider(config.YouverifyConfig{})
	endpoint, _ := p.resolveEndpoint(internal.ProviderRequest{
		CountryCode: "GH",
		DocType:     "PASSPORT",
	})
	if endpoint != "" {
		t.Errorf("expected empty for unsupported doc in GH, got %s", endpoint)
	}
}

func TestYouverifyProvider_ResolveEndpoint_NG_UnsupportedDoc(t *testing.T) {
	p := NewYouverifyProvider(config.YouverifyConfig{})
	endpoint, _ := p.resolveEndpoint(internal.ProviderRequest{
		CountryCode: "NG",
		DocType:     "PASSPORT",
	})
	if endpoint != "" {
		t.Errorf("expected empty for unsupported doc in NG, got %s", endpoint)
	}
}
