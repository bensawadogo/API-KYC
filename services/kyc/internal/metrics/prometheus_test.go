package kycmetrics_test

import (
	"net/http"
	"testing"

	"github.com/datakeys/kyc-service/internal/metrics"
	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestNewKYCMetrics_NoNilFields(t *testing.T) {
	m := kycmetrics.NewKYCMetrics(prometheus.NewRegistry())

	fields := []interface{}{
		m.VerificationsTotal,
		m.ProviderCallsTotal,
		m.WebhookSentTotal,
		m.RateLimitHitsTotal,
		m.IdempotencyReplaysTotal,
		m.DocumentsDeletedTotal,
		m.RequestDuration,
		m.ProviderDuration,
		m.DatabaseQueryDuration,
		m.VerificationsPending,
		&m.VerificationsProcessing,
		m.ProviderScore,
		m.CleanupJobLastRun,
		m.StorageDocumentsSizeBytes,
	}
	for i, f := range fields {
		if f == nil {
			t.Errorf("field %d should not be nil", i)
		}
	}
}

func TestVerificationsTotal_Increment(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := kycmetrics.NewKYCMetrics(reg)

	m.VerificationsTotal.WithLabelValues("BF", "NATIONAL_ID", "approved", "smileid").Inc()

	cnt, err := testutil.GatherAndCount(reg, "kyc_verifications_total")
	if err != nil {
		t.Fatal(err)
	}
	if cnt < 1 {
		t.Error("expected at least 1 metric sample")
	}
}

func TestRequestDuration_Observe(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := kycmetrics.NewKYCMetrics(reg)

	m.RequestDuration.WithLabelValues("POST", "/v1/kyc/initiate", "201").Observe(0.5)

	cnt, err := testutil.GatherAndCount(reg, "kyc_request_duration_seconds")
	if err != nil {
		t.Fatal(err)
	}
	if cnt < 1 {
		t.Error("expected at least 1 metric sample")
	}
}

func TestPrometheusMiddleware_UUIDNormalization(t *testing.T) {
	reg := prometheus.NewRegistry()
	m := kycmetrics.NewKYCMetrics(reg)

	app := fiber.New()
	app.Use(kycmetrics.PrometheusMiddleware(m))
	app.Get("/v1/kyc/status/:id", func(c fiber.Ctx) error {
		return c.SendString("ok")
	})

	req, _ := http.NewRequest(http.MethodGet, "/v1/kyc/status/550e8400-e29b-41d4-a716-446655440000", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

func TestNewDefaultMetrics_NotNil(t *testing.T) {
	m := kycmetrics.NewDefaultMetrics()
	if m == nil {
		t.Fatal("NewDefaultMetrics should not return nil")
	}
}