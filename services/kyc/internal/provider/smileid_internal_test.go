package provider

import (
	"testing"
)

func TestMapDocType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"NATIONAL_ID", "NATIONAL_ID"},
		{"PASSPORT", "PASSPORT"},
		{"DRIVERS_LICENSE", "DRIVERS_LICENSE"},
		{"VOTER_CARD", "VOTER_ID"},
		{"RESIDENCE_PERMIT", "RESIDENCE_PERMIT"},
		{"", ""},
	}
	for _, tt := range tests {
		result := mapDocType(tt.input)
		if result != tt.expected {
			t.Errorf("mapDocType(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

func TestIsSelfieData_WithSelfie(t *testing.T) {
	data := []byte(`{"selfie":"base64data","other":"value"}`)
	if !isSelfieData(data) {
		t.Error("expected true for data with selfie key")
	}
}

func TestIsSelfieData_WithoutSelfie(t *testing.T) {
	data := []byte(`{"document":"base64data"}`)
	if isSelfieData(data) {
		t.Error("expected false for data without selfie key")
	}
}

func TestIsSelfieData_InvalidJSON(t *testing.T) {
	data := []byte(`not-json`)
	if isSelfieData(data) {
		t.Error("expected false for invalid json")
	}
}

func TestIsSelfieData_Nil(t *testing.T) {
	if isSelfieData(nil) {
		t.Error("expected false for nil")
	}
}

func TestParseSmileIDActions_Nil(t *testing.T) {
	flags := parseSmileIDActions(nil)
	if flags != nil {
		t.Errorf("expected nil, got %v", flags)
	}
}

func TestParseSmileIDActions_Empty(t *testing.T) {
	flags := parseSmileIDActions(map[string]interface{}{})
	if len(flags) != 0 {
		t.Errorf("expected empty, got %v", flags)
	}
}

func TestParseSmileIDActions_Expired(t *testing.T) {
	actions := map[string]interface{}{
		"expired": "Failed",
	}
	flags := parseSmileIDActions(actions)
	if len(flags) != 1 || flags[0] != "EXPIRED_DOC" {
		t.Errorf("expected EXPIRED_DOC, got %v", flags)
	}
}

func TestParseSmileIDActions_Sanctions(t *testing.T) {
	actions := map[string]interface{}{
		"sanctions": "rejected",
	}
	flags := parseSmileIDActions(actions)
	if len(flags) != 1 || flags[0] != "SANCTIONS_MATCH" {
		t.Errorf("expected SANCTIONS_MATCH, got %v", flags)
	}
}

func TestParseSmileIDActions_InvalidFormat(t *testing.T) {
	actions := map[string]interface{}{
		"format": "failed",
	}
	flags := parseSmileIDActions(actions)
	if len(flags) != 1 || flags[0] != "INVALID_FORMAT" {
		t.Errorf("expected INVALID_FORMAT, got %v", flags)
	}
}

func TestParseSmileIDActions_IdFormat(t *testing.T) {
	actions := map[string]interface{}{
		"id_format": "Failed",
	}
	flags := parseSmileIDActions(actions)
	if len(flags) != 1 || flags[0] != "INVALID_FORMAT" {
		t.Errorf("expected INVALID_FORMAT, got %v", flags)
	}
}

func TestParseSmileIDActions_OtherFailed(t *testing.T) {
	actions := map[string]interface{}{
		"something": "failed",
	}
	flags := parseSmileIDActions(actions)
	if len(flags) != 1 || flags[0] != "LOW_CONFIDENCE" {
		t.Errorf("expected LOW_CONFIDENCE, got %v", flags)
	}
}

func TestParseSmileIDActions_NonStringValue(t *testing.T) {
	actions := map[string]interface{}{
		"expired": 123,
	}
	flags := parseSmileIDActions(actions)
	if len(flags) != 0 {
		t.Errorf("expected no flags for non-string value, got %v", flags)
	}
}

func TestParseSmileIDActions_Multiple(t *testing.T) {
	actions := map[string]interface{}{
		"expired":   "Failed",
		"sanctions": "rejected",
		"format":    "failed",
	}
	flags := parseSmileIDActions(actions)
	if len(flags) != 3 {
		t.Errorf("expected 3 flags, got %d: %v", len(flags), flags)
	}
}
