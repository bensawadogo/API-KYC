package middleware_test

import (
	"net/http"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/datakeys/kyc-service/internal/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func setupID(t *testing.T) (*middleware.IdempotencyStore, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	logger, _ := zap.NewDevelopment()
	im := middleware.NewIdempotencyStore(rdb, logger)
	t.Cleanup(func() { mr.Close() })
	return im, mr
}

func validUUID() string {
	return "550e8400-e29b-41d4-a716-446655440000"
}

func TestIdempotency_RedisDown_FailOpen(t *testing.T) {
	im, mr := setupID(t)
	app := fiber.New()
	app.Post("/initiate", im.Middleware, func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusCreated)
	})

	mr.Close()

	req, _ := http.NewRequest(http.MethodPost, "/initiate", nil)
	req.Header.Set("Idempotency-Key", validUUID())
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201 (fail-open), got %d", resp.StatusCode)
	}
}

func TestIdempotency_NoKey_Passes(t *testing.T) {
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

func TestIdempotency_ValidKey_FirstRequest_Passes(t *testing.T) {
	im, _ := setupID(t)
	callCount := 0
	app := fiber.New()
	app.Post("/initiate", im.Middleware, func(c fiber.Ctx) error {
		callCount++
		return c.SendStatus(fiber.StatusCreated)
	})

	req, _ := http.NewRequest(http.MethodPost, "/initiate", nil)
	req.Header.Set("Idempotency-Key", validUUID())
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
	if callCount != 1 {
		t.Errorf("handler should be called once, got %d", callCount)
	}
}

func TestIdempotency_ErrorResponse_NotCached(t *testing.T) {
	im, _ := setupID(t)
	callCount := 0
	app := fiber.New()
	app.Post("/initiate", im.Middleware, func(c fiber.Ctx) error {
		callCount++
		return c.SendStatus(fiber.StatusInternalServerError)
	})

	key := validUUID()

	req1, _ := http.NewRequest(http.MethodPost, "/initiate", nil)
	req1.Header.Set("Idempotency-Key", key)
	app.Test(req1)

	req2, _ := http.NewRequest(http.MethodPost, "/initiate", nil)
	req2.Header.Set("Idempotency-Key", key)
	app.Test(req2)

	if callCount != 2 {
		t.Errorf("handler should be called twice for errors, got %d", callCount)
	}
}
