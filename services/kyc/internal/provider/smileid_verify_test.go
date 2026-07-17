package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/datakeys/kyc-service/internal"
)

func TestSmileIDProvider_Verify_Approved(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		resp := map[string]interface{}{
			"confidence_value": float64(95),
			"Actions":          map[string]interface{}{},
			"Result":           "Verified",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &SmileIDProvider{
		client:      &http.Client{},
		testBaseURL: server.URL + "/",
	}

	result, err := p.Verify(context.Background(), internal.ProviderRequest{
		VerificationID: "v1",
		CountryCode:    "BF",
		DocType:        "NATIONAL_ID",
		DocNumber:      "B1234567",
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
	if result.Provider != "smileid" {
		t.Errorf("expected smileid, got %s", result.Provider)
	}
}

func TestSmileIDProvider_Verify_Rejected(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"confidence_value": float64(30),
			"Actions": map[string]interface{}{
				"expired": "Failed",
			},
			"Result": "Failed",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &SmileIDProvider{
		client:      &http.Client{},
		testBaseURL: server.URL + "/",
	}

	result, err := p.Verify(context.Background(), internal.ProviderRequest{
		VerificationID: "v1",
		CountryCode:    "BF",
		DocType:        "NATIONAL_ID",
		DocNumber:      "B1234567",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Approved {
		t.Error("expected not approved")
	}
	if result.Score != 0.30 {
		t.Errorf("expected 0.30, got %f", result.Score)
	}
	foundExpired := false
	for _, f := range result.Flags {
		if f == "EXPIRED_DOC" {
			foundExpired = true
			break
		}
	}
	if !foundExpired {
		t.Error("expected EXPIRED_DOC flag")
	}
}

func TestSmileIDProvider_Verify_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal"}`))
	}))
	defer server.Close()

	p := &SmileIDProvider{
		client:      &http.Client{},
		testBaseURL: server.URL + "/",
	}

	_, err := p.Verify(context.Background(), internal.ProviderRequest{
		VerificationID: "v1",
		CountryCode:    "BF",
		DocType:        "NATIONAL_ID",
	})
	if err == nil {
		t.Fatal("expected error for API error")
	}
}

func TestSmileIDProvider_Verify_SelfieJobType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		json.NewDecoder(r.Body).Decode(&payload)
		if payload["job_type"] != float64(2) {
			t.Errorf("expected job_type 2 for selfie, got %v", payload["job_type"])
		}
		resp := map[string]interface{}{
			"confidence_value": float64(90),
			"Actions":          map[string]interface{}{},
			"Result":           "Verified",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &SmileIDProvider{
		client:      &http.Client{},
		testBaseURL: server.URL + "/",
	}

	_, err := p.Verify(context.Background(), internal.ProviderRequest{
		VerificationID: "v1",
		CountryCode:    "BF",
		DocType:        "NATIONAL_ID",
		DocData:        []byte(`{"selfie":"image-data"}`),
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSmileIDProvider_Verify_LowConfidenceFlag(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"confidence_value": float64(65),
			"Actions":          map[string]interface{}{},
			"Result":           "Verified",
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	p := &SmileIDProvider{
		client:      &http.Client{},
		testBaseURL: server.URL + "/",
	}

	result, err := p.Verify(context.Background(), internal.ProviderRequest{
		VerificationID: "v1",
		CountryCode:    "BF",
		DocType:        "NATIONAL_ID",
		DocNumber:      "B1234567",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Approved {
		t.Error("expected not approved for low confidence")
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

func TestSmileIDProvider_Verify_NetworkError(t *testing.T) {
	p := &SmileIDProvider{
		client:      &http.Client{},
		testBaseURL: "http://127.0.0.1:1/",
	}

	_, err := p.Verify(context.Background(), internal.ProviderRequest{
		VerificationID: "v1",
		CountryCode:    "BF",
		DocType:        "NATIONAL_ID",
	})
	if err == nil {
		t.Fatal("expected error for network error")
	}
}
