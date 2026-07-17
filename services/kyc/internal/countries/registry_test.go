package countries_test

import (
	"testing"

	"github.com/datakeys/kyc-service/internal/countries"
)

func TestGetCountry_BurkinaFaso(t *testing.T) {
	c, found := countries.GetCountry("BF")
	if !found {
		t.Fatal("BF should be found")
	}
	if c.PhonePrefix != "+226" {
		t.Errorf("expected +226, got %s", c.PhonePrefix)
	}
	if c.Region != "WEST_AFRICA" {
		t.Errorf("expected WEST_AFRICA, got %s", c.Region)
	}
	if c.Name == "" {
		t.Error("name should not be empty")
	}
}

func TestGetCountry_Nigeria(t *testing.T) {
	c, found := countries.GetCountry("NG")
	if !found {
		t.Fatal("NG should be found")
	}
	if c.Provider == "" {
		t.Error("provider should not be empty")
	}
}

func TestGetCountry_Morocco(t *testing.T) {
	c, found := countries.GetCountry("MA")
	if !found {
		t.Fatal("MA should be found")
	}
	if c.Region != "NORTH_AFRICA" {
		t.Errorf("expected NORTH_AFRICA, got %s", c.Region)
	}
}

func TestGetCountry_SouthAfrica(t *testing.T) {
	c, found := countries.GetCountry("ZA")
	if !found {
		t.Fatal("ZA should be found")
	}
	if c.Region != "SOUTHERN_AFRICA" {
		t.Errorf("expected SOUTHERN_AFRICA, got %s", c.Region)
	}
}

func TestGetCountry_Unknown(t *testing.T) {
	_, found := countries.GetCountry("XX")
	if found {
		t.Error("XX should not be found")
	}
}

func TestIsDocTypeValid_CNIB_BF(t *testing.T) {
	valid := countries.IsDocTypeValid("BF", "NATIONAL_ID")
	if !valid {
		t.Error("NATIONAL_ID should be valid for BF")
	}
}

func TestIsDocTypeValid_Passport_BF(t *testing.T) {
	valid := countries.IsDocTypeValid("BF", "PASSPORT")
	if !valid {
		t.Error("PASSPORT should be valid for BF")
	}
}

func TestIsDocTypeValid_UnsupportedDoc(t *testing.T) {
	valid := countries.IsDocTypeValid("BF", "VOTER_CARD")
	if valid {
		t.Error("VOTER_CARD should not be valid for BF")
	}
}

func TestIsDocTypeValid_UnknownCountry(t *testing.T) {
	valid := countries.IsDocTypeValid("XX", "PASSPORT")
	if valid {
		t.Error("XX should not have valid doc types")
	}
}

func TestValidateDocNumber_CNIB_Valid(t *testing.T) {
	valid := countries.ValidateDocNumber("BF", "NATIONAL_ID", "B1234567")
	if !valid {
		t.Error("B1234567 should be a valid CNIB")
	}
}

func TestValidateDocNumber_CNIB_Invalid(t *testing.T) {
	valid := countries.ValidateDocNumber("BF", "NATIONAL_ID", "12345")
	if valid {
		t.Error("12345 should be invalid CNIB format")
	}
}

func TestValidateDocNumber_NIN_Nigeria(t *testing.T) {
	valid := countries.ValidateDocNumber("NG", "NATIONAL_ID", "12345678901")
	if !valid {
		t.Error("12345678901 should be a valid NIN")
	}
}

func TestValidateDocNumber_SouthAfrica(t *testing.T) {
	valid := countries.ValidateDocNumber("ZA", "NATIONAL_ID", "9001015009087")
	if !valid {
		t.Error("9001015009087 should be a valid SA ID")
	}
}

func TestValidateDocNumber_Passport_NoPattern(t *testing.T) {
	valid := countries.ValidateDocNumber("BF", "PASSPORT", "AB123456")
	if !valid {
		t.Error("passport without pattern should accept any number")
	}
}

func TestListCountries_Coverage(t *testing.T) {
	all := countries.ListCountries()
	if len(all) < 20 {
		t.Errorf("expected at least 20 countries, got %d", len(all))
	}

	expected := map[string]bool{"BF": false, "NG": false, "KE": false, "MA": false, "ZA": false}
	for _, c := range all {
		if _, ok := expected[c.Code]; ok {
			expected[c.Code] = true
		}
		if len(c.DocTypes) < 1 {
			t.Errorf("country %s has no document types", c.Code)
		}
	}
	for code, found := range expected {
		if !found {
			t.Errorf("expected country %s not found in list", code)
		}
	}
}