package countries

import (
	"testing"
)

func TestGetProvider_Found(t *testing.T) {
	provider := GetProvider("BF")
	if provider != "smileid" {
		t.Errorf("expected smileid, got %s", provider)
	}
}

func TestGetProvider_NotFound(t *testing.T) {
	provider := GetProvider("ZZ")
	if provider == "" {
		t.Error("expected non-empty fallback provider for unknown country")
	}
}

func TestGetPhonePrefix_Found(t *testing.T) {
	prefix := GetPhonePrefix("BF")
	if prefix != "+226" {
		t.Errorf("expected +226, got %s", prefix)
	}
}

func TestGetPhonePrefix_NotFound(t *testing.T) {
	prefix := GetPhonePrefix("ZZ")
	if prefix != "" {
		t.Errorf("expected empty, got %s", prefix)
	}
}

func TestGetPhonePrefix_Nigeria(t *testing.T) {
	prefix := GetPhonePrefix("NG")
	if prefix != "+234" {
		t.Errorf("expected +234, got %s", prefix)
	}
}

func TestListByRegion(t *testing.T) {
	west := ListByRegion("WEST_AFRICA")
	if len(west) == 0 {
		t.Fatal("expected at least one WEST_AFRICA country")
	}
	hasBF := false
	for _, c := range west {
		if c.Code == "BF" {
			hasBF = true
			break
		}
	}
	if !hasBF {
		t.Error("expected BF in WEST_AFRICA")
	}
}

func TestListByRegion_Empty(t *testing.T) {
	results := ListByRegion("NONEXISTENT")
	if len(results) != 0 {
		t.Errorf("expected empty, got %d", len(results))
	}
}

func TestListByRegion_CaseInsensitive(t *testing.T) {
	results := ListByRegion("west_africa")
	if len(results) == 0 {
		t.Error("expected results for lowercase west_africa (case-insensitive)")
	}
}

func TestAllAfricanCountryCodes(t *testing.T) {
	codes := AllAfricanCountryCodes()
	if len(codes) == 0 {
		t.Fatal("expected non-empty list")
	}
	hasBF := false
	for _, c := range codes {
		if c == "BF" {
			hasBF = true
			break
		}
	}
	if !hasBF {
		t.Error("expected BF in all african codes")
	}
}

func TestAllAfricanCountryCodes_Unique(t *testing.T) {
	codes := AllAfricanCountryCodes()
	seen := make(map[string]bool)
	for _, c := range codes {
		if seen[c] {
			t.Errorf("duplicate code: %s", c)
		}
		seen[c] = true
	}
}

func TestListCountries_Size(t *testing.T) {
	countries := ListCountries()
	if len(countries) < 40 {
		t.Errorf("expected at least 40 countries, got %d", len(countries))
	}
}

func TestGetCountry_Senegal(t *testing.T) {
	country, ok := GetCountry("SN")
	if !ok {
		t.Fatal("expected SN to be found")
	}
	if country.PhonePrefix != "+221" {
		t.Errorf("expected +221, got %s", country.PhonePrefix)
	}
}
