package provider

import (
	"context"
	"testing"

	"github.com/datakeys/kyc-service/internal"
)

func TestLocalProvider_IsFallbackCandidate_Known(t *testing.T) {
	p := NewLocalProvider()
	if !p.IsFallbackCandidate("BF") {
		t.Error("expected true for known country BF")
	}
}

func TestLocalProvider_IsFallbackCandidate_Unknown(t *testing.T) {
	p := NewLocalProvider()
	if p.IsFallbackCandidate("ZZ") {
		t.Error("expected false for unknown country")
	}
}

func TestLocalProvider_IsFallbackCandidate_CaseInsensitive(t *testing.T) {
	p := NewLocalProvider()
	if !p.IsFallbackCandidate("bf") {
		t.Error("expected true for lowercase bf")
	}
}

func TestLocalProvider_Verify_Production_InvalidDocNumber(t *testing.T) {
	p := NewLocalProviderWithSandbox(false)
	result, err := p.Verify(context.Background(), internal.ProviderRequest{
		CountryCode: "BF",
		DocType:     "NATIONAL_ID",
		DocNumber:   "INVALID",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Approved {
		t.Error("expected not approved for invalid doc number")
	}
}
