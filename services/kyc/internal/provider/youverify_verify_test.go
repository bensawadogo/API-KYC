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

func TestYouverifyProvider_Verify_Found(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"matchStatus": "found",
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &YouverifyProvider{
		cfg:    config.YouverifyConfig{ApiKey: "test", BaseURL: server.URL + "/"},
		client: &http.Client{},
	}

	result, err := p.Verify(context.Background(), internal.ProviderRequest{
		VerificationID: "v1",
		CountryCode:    "NG",
		DocType:        "NATIONAL_ID",
		DocNumber:      "12345678901",
	})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Approved {
		t.Error("expected approved")
	}
	if result.Score != 0.95 {
		t.Errorf("expected 0.95, got %f", result.Score)
	}
}

func TestYouverifyProvider_Verify_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"success": true,
			"data": map[string]interface{}{
				"matchStatus": "not_found",
			},
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &YouverifyProvider{
		cfg:    config.YouverifyConfig{ApiKey: "test", BaseURL: server.URL + "/"},
		client: &http.Client{},
	}

	result, err := p.Verify(context.Background(), internal.ProviderRequest{
		VerificationID: "v1",
		CountryCode:    "NG",
		DocType:        "NATIONAL_ID",
		DocNumber:      "00000000000",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Approved {
		t.Error("expected not approved")
	}
	if result.Score != 0.1 {
		t.Errorf("expected 0.1, got %f", result.Score)
	}
}

func TestYouverifyProvider_Verify_UnsupportedDoc(t *testing.T) {
	p := &YouverifyProvider{
		cfg:    config.YouverifyConfig{ApiKey: "test"},
		client: &http.Client{},
	}

	result, err := p.Verify(context.Background(), internal.ProviderRequest{
		VerificationID: "v1",
		CountryCode:    "BF",
		DocType:        "NATIONAL_ID",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Approved {
		t.Error("expected not approved for unsupported country")
	}
	if result.Score != 0.1 {
		t.Errorf("expected 0.1, got %f", result.Score)
	}
	found := false
	for _, f := range result.Flags {
		if f == "UNSUPPORTED_DOC_TYPE" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected UNSUPPORTED_DOC_TYPE flag")
	}
}

func TestYouverifyProvider_Verify_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"error":"upstream"}`))
	}))
	defer server.Close()

	p := &YouverifyProvider{
		cfg:    config.YouverifyConfig{ApiKey: "test", BaseURL: server.URL + "/"},
		client: &http.Client{},
	}

	_, err := p.Verify(context.Background(), internal.ProviderRequest{
		VerificationID: "v1",
		CountryCode:    "NG",
		DocType:        "NATIONAL_ID",
		DocNumber:      "12345678901",
	})
	if err == nil {
		t.Fatal("expected error for API error")
	}
}
