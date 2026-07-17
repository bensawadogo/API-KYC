package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/datakeys/kyc-service/config"
	"github.com/datakeys/kyc-service/internal"
)

func TestSumSubProvider_Verify_Green(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		switch callCount {
		case 1:
			resp := map[string]interface{}{"id": "applicant-1"}
			json.NewEncoder(w).Encode(resp)
		case 2:
			w.WriteHeader(http.StatusOK)
		case 3:
			w.WriteHeader(http.StatusOK)
		case 4:
			resp := map[string]interface{}{
				"review": map[string]interface{}{
					"reviewResult": map[string]interface{}{
						"reviewAnswer": "GREEN",
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	p := &SumSubProvider{
		cfg:    config.SumSubConfig{AppToken: "token", SecretKey: "secret", BaseURL: server.URL},
		client: &http.Client{},
	}

	result, err := p.Verify(context.Background(), internal.ProviderRequest{
		VerificationID: "v1",
		CountryCode:    "MA",
		DocType:        "NATIONAL_ID",
		DocNumber:      "AB123456",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Approved {
		t.Error("expected approved")
	}
	if result.Score != 0.92 {
		t.Errorf("expected 0.92, got %f", result.Score)
	}
}

func TestSumSubProvider_Verify_Red(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		switch callCount {
		case 1:
			resp := map[string]interface{}{"id": "applicant-2"}
			json.NewEncoder(w).Encode(resp)
		case 2:
			w.WriteHeader(http.StatusOK)
		case 3:
			w.WriteHeader(http.StatusOK)
		case 4:
			resp := map[string]interface{}{
				"review": map[string]interface{}{
					"reviewResult": map[string]interface{}{
						"reviewAnswer": "RED",
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	p := &SumSubProvider{
		cfg:    config.SumSubConfig{AppToken: "token", SecretKey: "secret", BaseURL: server.URL},
		client: &http.Client{},
	}

	result, err := p.Verify(context.Background(), internal.ProviderRequest{
		VerificationID: "v1",
		CountryCode:    "MA",
		DocType:        "NATIONAL_ID",
		DocNumber:      "AB123456",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Approved {
		t.Error("expected not approved")
	}
}

func TestSumSubProvider_Verify_WithDocumentUpload(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		switch callCount {
		case 1:
			resp := map[string]interface{}{"id": "applicant-3"}
			json.NewEncoder(w).Encode(resp)
		case 2:
			w.WriteHeader(http.StatusOK)
		case 3:
			w.WriteHeader(http.StatusOK)
		case 4:
			resp := map[string]interface{}{
				"review": map[string]interface{}{
					"reviewResult": map[string]interface{}{
						"reviewAnswer": "GREEN",
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}
	}))
	defer server.Close()

	p := &SumSubProvider{
		cfg:    config.SumSubConfig{AppToken: "token", SecretKey: "secret", BaseURL: server.URL},
		client: &http.Client{},
	}

	result, err := p.Verify(context.Background(), internal.ProviderRequest{
		VerificationID: "v1",
		CountryCode:    "MA",
		DocType:        "NATIONAL_ID",
		DocNumber:      "AB123456",
		DocData:        []byte("fake-image-bytes"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Approved {
		t.Error("expected approved")
	}
}

func TestSumSubProvider_Verify_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{"id": "applicant-4"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &SumSubProvider{
		cfg:    config.SumSubConfig{AppToken: "token", SecretKey: "secret", BaseURL: server.URL},
		client: &http.Client{},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := p.Verify(ctx, internal.ProviderRequest{
		VerificationID: "v1",
		CountryCode:    "MA",
		DocType:        "NATIONAL_ID",
	})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestSumSubProvider_Verify_ApplicantCreationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer server.Close()

	p := &SumSubProvider{
		cfg:    config.SumSubConfig{AppToken: "token", SecretKey: "secret", BaseURL: server.URL},
		client: &http.Client{},
	}

	_, err := p.Verify(context.Background(), internal.ProviderRequest{
		VerificationID: "v1",
		CountryCode:    "MA",
		DocType:        "NATIONAL_ID",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSumSubProvider_Sign(t *testing.T) {
	p := &SumSubProvider{
		cfg: config.SumSubConfig{SecretKey: "test-secret"},
	}

	sig := p.sign("POST", "/test", "1234567890", []byte(`{"key":"value"}`))
	if sig == "" {
		t.Error("expected non-empty signature")
	}
}

func TestMapSumSubDocType_AllTypes(t *testing.T) {
	tests := []struct{ in, out string }{
		{"NATIONAL_ID", "ID_CARD"},
		{"PASSPORT", "PASSPORT"},
		{"DRIVERS_LICENSE", "DRIVERS"},
		{"RESIDENCE_PERMIT", "RESIDENCE_PERMIT"},
		{"VOTER_CARD", "VOTER_CARD"},
		{"UNKNOWN", "UNKNOWN"},
		{"", ""},
	}
	for _, tt := range tests {
		r := mapSumSubDocType(tt.in)
		if r != tt.out {
			t.Errorf("mapSumSubDocType(%q) = %q, want %q", tt.in, r, tt.out)
		}
	}
}
