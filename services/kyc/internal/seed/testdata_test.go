package seed_test

import (
	"testing"

	"github.com/datakeys/kyc-service/internal/seed"
)

func TestSandboxProfiles_Count(t *testing.T) {
	if n := len(seed.SandboxProfiles); n != 8 {
		t.Errorf("expected 8 profiles, got %d", n)
	}
}

func TestSandboxProfiles_AllHavePhone(t *testing.T) {
	for i, p := range seed.SandboxProfiles {
		if p.Phone == "" {
			t.Errorf("profile %d (%s) has empty phone", i, p.FullName)
		}
	}
}

func TestSandboxProfiles_AllHaveDocNumber(t *testing.T) {
	for i, p := range seed.SandboxProfiles {
		if p.DocNumber == "" {
			t.Errorf("profile %d (%s) has empty doc_number", i, p.FullName)
		}
	}
}

func TestGetTestAPIKey_Format(t *testing.T) {
	key := seed.GetTestAPIKey()
	if len(key) < 20 {
		t.Errorf("api key too short: %s", key)
	}
}

func TestCurlExample_Format(t *testing.T) {
	profile := seed.SandboxProfiles[0]
	curl := seed.CurlExample(profile, "http://localhost:8081", "dk_test_key")
	if curl == "" {
		t.Fatal("curl example should not be empty")
	}
	if len(curl) < 50 {
		t.Errorf("curl example too short: %s", curl)
	}
}
