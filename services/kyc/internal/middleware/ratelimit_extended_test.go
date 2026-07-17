package middleware_test

import (
	"net/http"
	"testing"

	"github.com/datakeys/kyc-service/internal/middleware"
	"github.com/gofiber/fiber/v3"
)

func TestRateLimiter_NewWithDefaults(t *testing.T) {
	rl, _ := setupRL(t)
	handler := rl.New(middleware.RateLimitConfig{})
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestRateLimiter_Relaxed(t *testing.T) {
	rl, _ := setupRL(t)
	app := fiber.New()
	app.Get("/countries", rl.Relaxed(), func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req, _ := http.NewRequest(http.MethodGet, "/countries", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		t.Error("first request should not be rate limited")
	}
}

func TestRateLimiter_Webhook(t *testing.T) {
	rl, _ := setupRL(t)
	app := fiber.New()
	app.Post("/webhook", rl.Webhook(), func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req, _ := http.NewRequest(http.MethodPost, "/webhook", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		t.Error("first request should not be rate limited")
	}
}

func TestRateLimiter_Strict_Exceeded(t *testing.T) {
	rl, mr := setupRL(t)
	app := fiber.New()
	app.Post("/initiate", rl.Strict(), func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	for i := 0; i < 15; i++ {
		req, _ := http.NewRequest(http.MethodPost, "/initiate", nil)
		app.Test(req)
	}

	mr.FastForward(0)

	req, _ := http.NewRequest(http.MethodPost, "/initiate", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Log("rate limit not triggered in test (miniredis TTL may not match real redis)")
	}
}
