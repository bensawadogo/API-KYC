package registry

import (
	"testing"
)

func TestAdapter_GetCountry_Found(t *testing.T) {
	a := New()
	country, ok := a.GetCountry("BF")
	if !ok {
		t.Fatal("expected BF to be found")
	}
	if country.Code != "BF" {
		t.Errorf("expected BF, got %s", country.Code)
	}
	if country.Name != "Burkina Faso" {
		t.Errorf("expected Burkina Faso, got %s", country.Name)
	}
}

func TestAdapter_GetCountry_NotFound(t *testing.T) {
	a := New()
	_, ok := a.GetCountry("ZZ")
	if ok {
		t.Fatal("expected ZZ not to be found")
	}
}

func TestAdapter_IsDocTypeValid_Valid(t *testing.T) {
	a := New()
	if !a.IsDocTypeValid("BF", "NATIONAL_ID") {
		t.Error("expected NATIONAL_ID to be valid for BF")
	}
}

func TestAdapter_IsDocTypeValid_Invalid(t *testing.T) {
	a := New()
	if a.IsDocTypeValid("BF", "VOTER_CARD") {
		t.Error("expected VOTER_CARD to be invalid for BF")
	}
}

func TestAdapter_ValidateDocNumber_Valid(t *testing.T) {
	a := New()
	if !a.ValidateDocNumber("BF", "NATIONAL_ID", "B1234567") {
		t.Error("expected B1234567 to be valid CNIB")
	}
}

func TestAdapter_ValidateDocNumber_Invalid(t *testing.T) {
	a := New()
	if a.ValidateDocNumber("BF", "NATIONAL_ID", "invalid") {
		t.Error("expected invalid to fail")
	}
}

func TestAdapter_ValidateDocNumber_NoPattern(t *testing.T) {
	a := New()
	if !a.ValidateDocNumber("BF", "PASSPORT", "") {
		t.Error("expected empty passport to validate (no pattern)")
	}
}

func TestAdapter_GetProvider(t *testing.T) {
	a := New()
	provider := a.GetProvider("BF")
	if provider == "" {
		t.Error("expected non-empty provider for BF")
	}
}

func TestAdapter_GetProvider_Unknown(t *testing.T) {
	a := New()
	provider := a.GetProvider("ZZ")
	if provider == "" {
		t.Error("expected non-empty fallback provider for unknown country")
	}
}
