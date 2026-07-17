package handler_test

import (
	"net/http"
	"testing"

	"github.com/datakeys/kyc-service/internal/handler"
	"github.com/gofiber/fiber/v3"
)

func TestLiveness(t *testing.T) {
	app := fiber.New()
	h := handler.NewHealthHandler(nil, nil, nil, nil)
	app.Get("/health/live", h.Live)

	req, _ := http.NewRequest(http.MethodGet, "/health/live", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestReadiness(t *testing.T) {
	app := fiber.New()
	h := handler.NewHealthHandler(nil, nil, nil, nil)
	app.Get("/health/ready", h.Ready)

	req, _ := http.NewRequest(http.MethodGet, "/health/ready", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestNewHealthHandler(t *testing.T) {
	h := handler.NewHealthHandler(nil, nil, nil, nil)
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}
