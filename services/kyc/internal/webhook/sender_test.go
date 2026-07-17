package webhook

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/datakeys/kyc-service/internal/model"
)

func TestNewHTTPSender(t *testing.T) {
	s := NewHTTPSender()
	if s == nil {
		t.Fatal("expected non-nil sender")
	}
	if s.client.Timeout != 15*time.Second {
		t.Errorf("expected 15s timeout, got %v", s.client.Timeout)
	}
}

func TestSend_Success(t *testing.T) {
	var receivedPayload model.WebhookPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected application/json, got %s", r.Header.Get("Content-Type"))
		}
		if err := json.NewDecoder(r.Body).Decode(&receivedPayload); err != nil {
			t.Errorf("decode payload: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewHTTPSender()
	payload := &model.WebhookPayload{
		Event: "kyc.verification.completed",
		Data: model.VerificationResult{
			VerificationID: "v1",
			Status:         "approved",
			Provider:       "local",
		},
	}

	err := sender.Send(context.Background(), server.URL, payload)
	if err != nil {
		t.Fatal(err)
	}

	if receivedPayload.Event != "kyc.verification.completed" {
		t.Errorf("expected kyc.verification.completed, got %s", receivedPayload.Event)
	}
	if receivedPayload.Data.VerificationID != "v1" {
		t.Errorf("expected v1, got %s", receivedPayload.Data.VerificationID)
	}
}

func TestSend_WithSignature(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if sig := r.Header.Get("X-KYC-Signature"); sig != "test-sig" {
			t.Errorf("expected test-sig, got %s", sig)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewHTTPSender()
	payload := &model.WebhookPayload{
		Event:     "kyc.verification.completed",
		Signature: "test-sig",
		Data:      model.VerificationResult{VerificationID: "v2"},
	}

	err := sender.Send(context.Background(), server.URL, payload)
	if err != nil {
		t.Fatal(err)
	}
}

func TestSend_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	sender := NewHTTPSender()
	payload := &model.WebhookPayload{
		Event: "kyc.verification.completed",
		Data:  model.VerificationResult{VerificationID: "v3"},
	}

	err := sender.Send(context.Background(), server.URL, payload)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestSend_InvalidURL(t *testing.T) {
	sender := NewHTTPSender()
	payload := &model.WebhookPayload{
		Event: "kyc.verification.completed",
		Data:  model.VerificationResult{VerificationID: "v4"},
	}

	err := sender.Send(context.Background(), "http://invalid.local:99999/nonexistent", payload)
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestSend_ContextTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	sender := NewHTTPSender()
	payload := &model.WebhookPayload{
		Event: "kyc.verification.completed",
		Data:  model.VerificationResult{VerificationID: "v5"},
	}

	err := sender.Send(ctx, server.URL, payload)
	if err == nil {
		t.Fatal("expected context timeout error")
	}
}
