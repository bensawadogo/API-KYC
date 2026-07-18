package resilience_test

import (
	"context"
	"errors"
	"testing"

	"github.com/datakeys/kyc-service/internal"
	"github.com/datakeys/kyc-service/internal/resilience"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type mockProvider struct {
	name       string
	shouldFail bool
	callCount  int
}

func (m *mockProvider) Verify(_ context.Context, _ internal.ProviderRequest) (*internal.ProviderResult, error) {
	m.callCount++
	if m.shouldFail {
		return nil, errors.New("provider down")
	}
	return &internal.ProviderResult{
		Approved: true, Score: 0.95, Provider: m.name,
	}, nil
}

func (m *mockProvider) SupportedCountries() []string {
	return []string{"BF", "NG"}
}

func (m *mockProvider) Name() string { return m.name }

func TestFallbackRouter_UsesFirstProvider(t *testing.T) {
	primary := &mockProvider{name: "primary"}
	secondary := &mockProvider{name: "secondary"}

	rp1 := resilience.NewResilientProvider(primary,
		resilience.DefaultCBConfig("primary"),
		resilience.DefaultRetryConfig(),
		zap.NewNop())
	rp2 := resilience.NewResilientProvider(secondary,
		resilience.DefaultCBConfig("secondary"),
		resilience.DefaultRetryConfig(),
		zap.NewNop())

	router := resilience.NewFallbackRouter([]*resilience.ResilientProvider{rp1, rp2}, zap.NewNop())

	result, provider, err := router.Verify(context.Background(), internal.ProviderRequest{
		VerificationID: "test-uuid",
		CountryCode:    "BF",
	})

	assert.NoError(t, err)
	assert.Equal(t, "primary", provider)
	assert.True(t, result.Approved)
	assert.Equal(t, 1, primary.callCount)
	assert.Equal(t, 0, secondary.callCount)
}

func TestFallbackRouter_FallsBackWhenPrimaryFails(t *testing.T) {
	cfg := resilience.DefaultCBConfig("primary-fail")
	cfg.MaxFailures = 1

	failing := &mockProvider{name: "primary-fail", shouldFail: true}
	backup := &mockProvider{name: "backup"}

	rp1 := resilience.NewResilientProvider(failing, cfg,
		resilience.RetryConfig{MaxAttempts: 1}, zap.NewNop())
	rp2 := resilience.NewResilientProvider(backup,
		resilience.DefaultCBConfig("backup"),
		resilience.DefaultRetryConfig(), zap.NewNop())

	router := resilience.NewFallbackRouter([]*resilience.ResilientProvider{rp1, rp2}, zap.NewNop())

	router.Verify(context.Background(), internal.ProviderRequest{VerificationID: "first"})

	result, provider, err := router.Verify(context.Background(), internal.ProviderRequest{
		VerificationID: "second",
		CountryCode:    "BF",
	})

	assert.NoError(t, err)
	assert.Equal(t, "backup", provider)
	assert.True(t, result.Approved)
}

func TestFallbackRouter_ReturnsErrorWhenAllDown(t *testing.T) {
	cfg := resilience.DefaultCBConfig("all-fail")
	cfg.MaxFailures = 1

	p1 := &mockProvider{name: "p1", shouldFail: true}
	p2 := &mockProvider{name: "p2", shouldFail: true}

	rp1 := resilience.NewResilientProvider(p1, cfg,
		resilience.RetryConfig{MaxAttempts: 1}, zap.NewNop())
	rp2 := resilience.NewResilientProvider(p2, cfg,
		resilience.RetryConfig{MaxAttempts: 1}, zap.NewNop())

	router := resilience.NewFallbackRouter([]*resilience.ResilientProvider{rp1, rp2}, zap.NewNop())

	router.Verify(context.Background(), internal.ProviderRequest{VerificationID: "open-circuits"})

	_, _, err := router.Verify(context.Background(), internal.ProviderRequest{
		VerificationID: "all-down",
		CountryCode:    "BF",
	})

	assert.Error(t, err)
}

func TestFallbackRouter_ProviderStatuses(t *testing.T) {
	p := &mockProvider{name: "status-test"}
	rp := resilience.NewResilientProvider(p,
		resilience.DefaultCBConfig("status-test"),
		resilience.DefaultRetryConfig(), zap.NewNop())
	router := resilience.NewFallbackRouter([]*resilience.ResilientProvider{rp}, zap.NewNop())

	statuses := router.ProviderStatuses()
	assert.Contains(t, statuses, "status-test")
	assert.Equal(t, "closed", statuses["status-test"])
}
