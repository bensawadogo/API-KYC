package middleware_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIdempotency_NoKey_PassesThrough(t *testing.T) {
	app, _ := setupIdmpApp(t)
	req, err := http.NewRequest(http.MethodPost, "/initiate", strings.NewReader(`{}`))
	require.NoError(t, err)
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
}

func TestIdempotency_DuplicateKey_Replayed(t *testing.T) {
	app, _ := setupIdmpApp(t)
	key := "550e8400-e29b-41d4-a716-446655440000"

	req1, err := http.NewRequest(http.MethodPost, "/initiate", strings.NewReader(`{}`))
	require.NoError(t, err)
	req1.Header.Set("Idempotency-Key", key)
	resp1, err := app.Test(req1)
	require.NoError(t, err)
	t.Logf("first call status: %d", resp1.StatusCode)
	assert.Equal(t, 201, resp1.StatusCode)

	req2, err := http.NewRequest(http.MethodPost, "/initiate", strings.NewReader(`{}`))
	require.NoError(t, err)
	req2.Header.Set("Idempotency-Key", key)
	resp2, err := app.Test(req2)
	require.NoError(t, err)
	t.Logf("second call status: %d, replayed: %s", resp2.StatusCode, resp2.Header.Get("Idempotent-Replayed"))
	assert.Equal(t, 201, resp2.StatusCode)
	assert.Equal(t, "true", resp2.Header.Get("Idempotency-Replayed"))
}

func TestIdempotency_WithKey_Passes(t *testing.T) {
	app, _ := setupIdmpApp(t)
	req, err := http.NewRequest(http.MethodPost, "/initiate", strings.NewReader(`{}`))
	require.NoError(t, err)
	req.Header.Set("Idempotency-Key", "custom-key-123")
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 201, resp.StatusCode)
}
