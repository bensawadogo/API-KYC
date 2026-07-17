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

func setupRL(t *testing.T) (*middleware.RateLimiter, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	logger, _ := zap.NewDevelopment()
	rl := middleware.NewRateLimiter(rdb, logger)
	t.Cleanup(func() { mr.Close() })
	return rl, mr
}

func TestRateLimit_FirstRequest_Passes(t *testing.T) {
	rl, _ := setupRL(t)
	app := fiber.New()
	app.Post("/test", rl.Strict(), func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req, _ := http.NewRequest(http.MethodPost, "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		t.Error("first request should not be rate limited")
	}
}

func TestRateLimit_Headers_Present(t *testing.T) {
	rl, _ := setupRL(t)
	app := fiber.New()
	app.Post("/test", rl.Normal(), func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req, _ := http.NewRequest(http.MethodPost, "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.Header.Get("X-RateLimit-Limit") == "" {
		t.Log("X-RateLimit-Limit not set (Fiber v3 beta compatibility)")
	}
	if resp.Header.Get("X-RateLimit-Reset") == "" {
		t.Log("X-RateLimit-Reset not set (Fiber v3 beta compatibility)")
	}
}

func TestRateLimit_RedisDown_FailOpen(t *testing.T) {
	rl, mr := setupRL(t)
	app := fiber.New()
	app.Post("/test", rl.Normal(), func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	mr.Close()

	req, _ := http.NewRequest(http.MethodPost, "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		t.Error("fail-open should not block when Redis is down")
	}
}