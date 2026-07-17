package handler

import (
	"testing"
)

func TestSuccessResponse(t *testing.T) {
	resp := SuccessResponse(map[string]string{"key": "value"})
	if !resp.Success {
		t.Error("expected success true")
	}
	if resp.Error != nil {
		t.Errorf("expected nil error, got %v", *resp.Error)
	}
	if resp.Data == nil {
		t.Error("expected non-nil data")
	}
	if resp.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestSuccessResponse_NilData(t *testing.T) {
	resp := SuccessResponse(nil)
	if !resp.Success {
		t.Error("expected success true")
	}
}

func TestErrorResponse(t *testing.T) {
	resp := ErrorResponse("something went wrong")
	if resp.Success {
		t.Error("expected success false")
	}
	if resp.Data != nil {
		t.Error("expected nil data")
	}
	if resp.Error == nil {
		t.Fatal("expected non-nil error")
	}
	if *resp.Error != "something went wrong" {
		t.Errorf("expected 'something went wrong', got '%s'", *resp.Error)
	}
	if resp.Timestamp == "" {
		t.Error("expected non-empty timestamp")
	}
}

func TestErrorResponse_EmptyMessage(t *testing.T) {
	resp := ErrorResponse("")
	if resp.Error == nil {
		t.Fatal("expected non-nil error")
	}
	if *resp.Error != "" {
		t.Errorf("expected empty string, got '%s'", *resp.Error)
	}
}
