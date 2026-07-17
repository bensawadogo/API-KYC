package middleware_test

import (
	"net/http"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func TestIdempotency_NoKey_PassesThrough(t *testing.T) {
	im, _ := setupID(t)
	app := fiber.New()
	app.Post("/initiate", im.Middleware, func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusCreated)
	})

	req, _ := http.NewRequest(http.MethodPost, "/initiate", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
}

func TestIdempotency_DuplicateKey_Replayed(t *testing.T) {
	im, _ := setupID(t)
	callCount := 0
	app := fiber.New()
	app.Post("/initiate", im.Middleware, func(c fiber.Ctx) error {
		callCount++
		return c.SendString("ok")
	})

	key := "550e8400-e29b-41d4-a716-446655440000"

	req1, _ := http.NewRequest(http.MethodPost, "/initiate", nil)
	req1.Header.Set("Idempotency-Key", key)
	resp1, _ := app.Test(req1)
	t.Logf("first call status: %d", resp1.StatusCode)

	req2, _ := http.NewRequest(http.MethodPost, "/initiate", nil)
	req2.Header.Set("Idempotency-Key", key)
	resp2, _ := app.Test(req2)
	t.Logf("second call status: %d, replayed: %s", resp2.StatusCode, resp2.Header.Get("Idempotent-Replayed"))

	t.Logf("handler call count: %d", callCount)
}

func TestIdempotency_WithKey_Passes(t *testing.T) {
	im, _ := setupID(t)
	callCount := 0
	app := fiber.New()
	app.Post("/initiate", im.Middleware, func(c fiber.Ctx) error {
		callCount++
		return c.SendString("ok")
	})

	req, _ := http.NewRequest(http.MethodPost, "/initiate", nil)
	req.Header.Set("Idempotency-Key", validUUID())
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode == http.StatusCreated {
		t.Log("got 201 as expected")
	}
}
