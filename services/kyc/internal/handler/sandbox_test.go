package handler_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/datakeys/kyc-service/internal/handler"
	"github.com/datakeys/kyc-service/internal/seed"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

func TestSandboxHandler_Profiles(t *testing.T) {
	app := fiber.New()
	logger, _ := zap.NewDevelopment()
	h := handler.NewSandboxHandler(seed.SandboxProfiles, "http://localhost:8081", "dk_test_key", logger)
	app.Get("/sandbox/profiles", h.Profiles)

	req, _ := http.NewRequest(http.MethodGet, "/sandbox/profiles", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if body["success"] != true {
		t.Error("expected success true")
	}
	data, ok := body["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data object")
	}
	profiles, ok := data["profiles"].([]interface{})
	if !ok {
		t.Fatal("expected profiles array")
	}
	if len(profiles) != 8 {
		t.Errorf("expected 8 profiles, got %d", len(profiles))
	}
	if data["total_profiles"] != float64(8) {
		t.Errorf("expected total_profiles 8, got %v", data["total_profiles"])
	}
	first := profiles[0].(map[string]interface{})
	if first["curl_example"] == "" {
		t.Error("expected curl_example in first profile")
	}
}

func TestSandboxHandler_Reset(t *testing.T) {
	app := fiber.New()
	logger, _ := zap.NewDevelopment()
	h := handler.NewSandboxHandler(seed.SandboxProfiles, "http://localhost:8081", "dk_test_key", logger)
	app.Get("/sandbox/reset", h.Reset)

	req, _ := http.NewRequest(http.MethodGet, "/sandbox/reset", nil)
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestSandboxHandler_Simulate_Approved(t *testing.T) {
	app := fiber.New()
	logger, _ := zap.NewDevelopment()
	h := handler.NewSandboxHandler(seed.SandboxProfiles, "http://localhost:8081", "dk_test_key", logger)
	app.Post("/sandbox/simulate", h.Simulate)

	body := `{"scenario":"approved"}`
	req, _ := http.NewRequest(http.MethodPost, "/sandbox/simulate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestSandboxHandler_Simulate_UnknownScenario(t *testing.T) {
	app := fiber.New()
	logger, _ := zap.NewDevelopment()
	h := handler.NewSandboxHandler(seed.SandboxProfiles, "http://localhost:8081", "dk_test_key", logger)
	app.Post("/sandbox/simulate", h.Simulate)

	body := `{"scenario":"unknown"}`
	req, _ := http.NewRequest(http.MethodPost, "/sandbox/simulate", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, fiber.TestConfig{Timeout: 5 * time.Second})
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}
