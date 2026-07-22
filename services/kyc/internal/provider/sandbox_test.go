package provider_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/datakeys/kyc-service/internal"
	"github.com/datakeys/kyc-service/internal/model"
	"github.com/datakeys/kyc-service/internal/provider"
)

func TestSandboxProvider_Name(t *testing.T) {
	sp := provider.NewSandboxProvider()
	if n := sp.Name(); n != "sandbox" {
		t.Errorf("expected 'sandbox', got '%s'", n)
	}
}

func TestSandboxProvider_SupportedCountries(t *testing.T) {
	sp := provider.NewSandboxProvider()
	countries := sp.SupportedCountries()
	if len(countries) < 10 {
		t.Errorf("expected many african countries, got %d", len(countries))
	}
}

func TestSandboxProvider_Approved_EndsWith0000(t *testing.T) {
	sp := provider.NewSandboxProvider()
	result, err := sp.Verify(context.Background(), req("XXXX0000"))
	if err != nil {
		t.Fatal(err)
	}
	if !result.Approved {
		t.Error("expected approved")
	}
	if result.Score != 0.95 {
		t.Errorf("expected score 0.95, got %f", result.Score)
	}
}

func TestSandboxProvider_Rejected_EndsWith1111(t *testing.T) {
	sp := provider.NewSandboxProvider()
	result, err := sp.Verify(context.Background(), req("XXXX1111"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Approved {
		t.Error("expected rejected")
	}
	if result.Score != 0.15 {
		t.Errorf("expected score 0.15, got %f", result.Score)
	}
	hasFlag(t, result.Flags, model.FlagLowConfidence)
}

func TestSandboxProvider_Sanctions_EndsWith2222(t *testing.T) {
	sp := provider.NewSandboxProvider()
	result, err := sp.Verify(context.Background(), req("XXXX2222"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Approved {
		t.Error("expected rejected")
	}
	hasFlag(t, result.Flags, model.FlagSanctionsMatch)
}

func TestSandboxProvider_PEP_EndsWith3333(t *testing.T) {
	sp := provider.NewSandboxProvider()
	result, err := sp.Verify(context.Background(), req("XXXX3333"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Approved {
		t.Error("expected not approved")
	}
	hasFlag(t, result.Flags, model.FlagPEPDetected)
}

func TestSandboxProvider_ExpiredDoc_EndsWith4444(t *testing.T) {
	sp := provider.NewSandboxProvider()
	result, err := sp.Verify(context.Background(), req("XXXX4444"))
	if err != nil {
		t.Fatal(err)
	}
	if result.Approved {
		t.Error("expected rejected")
	}
	hasFlag(t, result.Flags, model.FlagExpiredDoc)
}

func TestSandboxProvider_Default_Approved(t *testing.T) {
	sp := provider.NewSandboxProvider()
	result, err := sp.Verify(context.Background(), req("ABC12345"))
	if err != nil {
		t.Fatal(err)
	}
	if !result.Approved {
		t.Error("expected approved for default case")
	}
	if result.Score != 0.92 {
		t.Errorf("expected score 0.92, got %f", result.Score)
	}
}

func TestSandboxProvider_ContextCancelled(t *testing.T) {
	sp := provider.NewSandboxProvider()
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()
	time.Sleep(10 * time.Millisecond)
	_, err := sp.Verify(ctx, req("XXXX0000"))
	if err == nil {
		t.Error("expected context cancellation error")
	}
}

func req(docNumber string) internal.ProviderRequest {
	return internal.ProviderRequest{
		VerificationID: "sandbox-test",
		CountryCode:    "BF",
		DocType:        "NATIONAL_ID",
		DocNumber:      docNumber,
		Phone:          "+22670000001",
	}
}

func hasFlag(t *testing.T, flags []string, flag string) {
	t.Helper()
	for _, f := range flags {
		if f == flag {
			return
		}
	}
	t.Errorf("expected flag %s in %v", flag, flags)
}

var _ = fmt.Sprintf
