package provider

import (
	"testing"
)

func TestMapSumSubReview_Green(t *testing.T) {
	result := mapSumSubReview("GREEN")
	if !result.Approved {
		t.Error("expected approved")
	}
	if result.Score != 0.92 {
		t.Errorf("expected 0.92, got %f", result.Score)
	}
}

func TestMapSumSubReview_Red(t *testing.T) {
	result := mapSumSubReview("RED")
	if result.Approved {
		t.Error("expected not approved")
	}
	if result.Score != 0.15 {
		t.Errorf("expected 0.15, got %f", result.Score)
	}
	found := false
	for _, f := range result.Flags {
		if f == "LOW_CONFIDENCE" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected LOW_CONFIDENCE flag")
	}
}

func TestMapSumSubReview_Unknown(t *testing.T) {
	result := mapSumSubReview("YELLOW")
	if result.Approved {
		t.Error("expected not approved for unknown")
	}
	if result.Score != 0.50 {
		t.Errorf("expected 0.50, got %f", result.Score)
	}
	found := false
	for _, f := range result.Flags {
		if f == "MANUAL_REVIEW_REQUIRED" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected MANUAL_REVIEW_REQUIRED flag")
	}
}

func TestMapSumSubReview_CaseInsensitive(t *testing.T) {
	green := mapSumSubReview("green")
	if !green.Approved {
		t.Error("expected approved for lowercase green")
	}
}

func TestMapSumSubDocType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"NATIONAL_ID", "ID_CARD"},
		{"PASSPORT", "PASSPORT"},
		{"DRIVERS_LICENSE", "DRIVERS"},
		{"RESIDENCE_PERMIT", "RESIDENCE_PERMIT"},
		{"VOTER_CARD", "VOTER_CARD"},
		{"", ""},
	}
	for _, tt := range tests {
		result := mapSumSubDocType(tt.input)
		if result != tt.expected {
			t.Errorf("mapSumSubDocType(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
