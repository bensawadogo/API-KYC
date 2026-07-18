package middleware_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/datakeys/kyc-service/internal/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupIdmpApp(t *testing.T) (*fiber.App, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	store := middleware.NewIdempotencyStore(rdb, zap.NewNop())

	app := fiber.New()
	app.Use(store.Middleware)
	app.Post("/initiate", func(c fiber.Ctx) error {
		return c.Status(201).JSON(fiber.Map{"id": "test-uuid"})
	})
	t.Cleanup(mr.Close)
	return app, mr
}

func TestIdempotency_NoKeyPassesThrough(t *testing.T) {
	app, _ := setupIdmpApp(t)
	req, err := http.NewRequest(http.MethodPost, "/initiate", strings.NewReader(`{"test":true}`))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
}

func TestIdempotency_FirstRequestSucceeds(t *testing.T) {
	app, _ := setupIdmpApp(t)
	req, err := http.NewRequest(http.MethodPost, "/initiate", strings.NewReader(`{}`))
	require.NoError(t, err)
	req.Header.Set("Idempotency-Key", "test-key-001")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
}

func TestIdempotency_DuplicateReturnsCache(t *testing.T) {
	app, _ := setupIdmpApp(t)
	key := "test-key-002"

	req1, err := http.NewRequest(http.MethodPost, "/initiate", strings.NewReader(`{}`))
	require.NoError(t, err)
	req1.Header.Set("Idempotency-Key", key)
	resp1, err := app.Test(req1)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp1.StatusCode)

	req2, err := http.NewRequest(http.MethodPost, "/initiate", strings.NewReader(`{}`))
	require.NoError(t, err)
	req2.Header.Set("Idempotency-Key", key)
	resp2, err := app.Test(req2)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp2.StatusCode)
	assert.Equal(t, "true", resp2.Header.Get("Idempotency-Replayed"))
}
