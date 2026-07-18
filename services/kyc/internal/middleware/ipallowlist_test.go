package middleware_test

import (
	"net/http"
	"testing"

	"github.com/datakeys/kyc-service/internal/middleware"
	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupIPApp(cidrs []string) *fiber.App {
	app := fiber.New()
	logger := zap.NewNop()
	al := middleware.NewIPAllowlist(cidrs, logger)
	app.Use(al.Middleware)
	app.Get("/webhook", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})
	return app
}

func TestIPAllowlist_AllowsWhenEmpty(t *testing.T) {
	app := setupIPApp([]string{})
	req, err := http.NewRequest(http.MethodGet, "/webhook", nil)
	require.NoError(t, err)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestIPAllowlist_AllowsMatchingIP(t *testing.T) {
	app := setupIPApp([]string{"0.0.0.0/0"})
	req, err := http.NewRequest(http.MethodGet, "/webhook", nil)
	require.NoError(t, err)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}

func TestIPAllowlist_BlocksNonMatchingIP(t *testing.T) {
	app := setupIPApp([]string{"10.0.0.0/8"})
	req, err := http.NewRequest(http.MethodGet, "/webhook", nil)
	require.NoError(t, err)
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 403, resp.StatusCode)
}

func TestIPAllowlist_AllowsPrivateRange(t *testing.T) {
	app := setupIPApp([]string{"10.0.0.0/8", "172.16.0.0/12"})
	req, err := http.NewRequest(http.MethodGet, "/webhook", nil)
	require.NoError(t, err)
	req.Header.Set("X-Forwarded-For", "10.5.3.2")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
}
