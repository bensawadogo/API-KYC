package handler_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/datakeys/kyc-service/internal/handler"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func TestSmileIDWebhook_Accepted(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := &mockService{}
	h := handler.NewKYCHandler(svc, logger, "", "", nil)

	app := fiber.New()
	app.Post("/webhook/smileid", h.SmileIDWebhook)

	body := map[string]interface{}{
		"user_id": "test-verification-id",
		"result":  "approved",
	}
	data, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/webhook/smileid", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Smile-Signature", "test-sig")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestSmileIDWebhook_MissingVerificationID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := &mockService{}
	h := handler.NewKYCHandler(svc, logger, "", "", nil)

	app := fiber.New()
	app.Post("/webhook/smileid", h.SmileIDWebhook)

	body := map[string]interface{}{"result": "approved"}
	data, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/webhook/smileid", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestSumSubWebhook_Accepted(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := &mockService{}
	h := handler.NewKYCHandler(svc, logger, "", "", nil)

	app := fiber.New()
	app.Post("/webhook/sumsub", h.SumSubWebhook)

	body := map[string]interface{}{
		"externalUserId": "test-verification-id",
		"reviewResult":   "GREEN",
	}
	data, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/webhook/sumsub", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Payload-Digest", "test-digest")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestSumSubWebhook_MissingVerificationID(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := &mockService{}
	h := handler.NewKYCHandler(svc, logger, "", "", nil)

	app := fiber.New()
	app.Post("/webhook/sumsub", h.SumSubWebhook)

	body := map[string]interface{}{"result": "GREEN"}
	data, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/webhook/sumsub", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestSmileIDWebhook_WithSignatureVerification(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := &mockService{}
	h := handler.NewKYCHandler(svc, logger, "test-secret-key", "", nil)

	app := fiber.New()
	app.Post("/webhook/smileid", h.SmileIDWebhook)

	body := map[string]interface{}{
		"user_id": "v123",
		"result":  "approved",
	}
	data, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/webhook/smileid", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Smile-Signature", "deadbeef")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 for bad signature, got %d", resp.StatusCode)
	}
}

func TestSumSubWebhook_WithSignatureVerification(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	svc := &mockService{}
	h := handler.NewKYCHandler(svc, logger, "", "test-secret-key", nil)

	app := fiber.New()
	app.Post("/webhook/sumsub", h.SumSubWebhook)

	body := map[string]interface{}{
		"externalUserId": "v123",
		"reviewResult":   "GREEN",
	}
	data, _ := json.Marshal(body)

	req, _ := http.NewRequest(http.MethodPost, "/webhook/sumsub", bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Payload-Digest", "deadbeef")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected 401 for bad signature, got %d", resp.StatusCode)
	}
}
