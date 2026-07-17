package kycmetrics

import (
	"regexp"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/prometheus/client_golang/prometheus"
)

var uuidRegex = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type KYCMetrics struct {
	VerificationsTotal        *prometheus.CounterVec
	ProviderCallsTotal        *prometheus.CounterVec
	WebhookSentTotal          *prometheus.CounterVec
	RateLimitHitsTotal        *prometheus.CounterVec
	IdempotencyReplaysTotal   *prometheus.CounterVec
	DocumentsDeletedTotal     *prometheus.CounterVec
	RequestDuration           *prometheus.HistogramVec
	ProviderDuration          *prometheus.HistogramVec
	DatabaseQueryDuration     *prometheus.HistogramVec
	VerificationsPending      *prometheus.GaugeVec
	VerificationsProcessing   prometheus.Gauge
	ProviderScore             *prometheus.SummaryVec
	CleanupJobLastRun         *prometheus.GaugeVec
	StorageDocumentsSizeBytes *prometheus.GaugeVec
}

func NewKYCMetrics(reg prometheus.Registerer) *KYCMetrics {
	m := &KYCMetrics{
		VerificationsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kyc_verifications_total",
				Help: "Nombre total de vérifications KYC par pays, type de document, statut et provider",
			},
			[]string{"country_code", "doc_type", "status", "provider"},
		),
		ProviderCallsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kyc_provider_calls_total",
				Help: "Nombre total d'appels aux providers d'identité externes",
			},
			[]string{"provider", "status"},
		),
		WebhookSentTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kyc_webhook_sent_total",
				Help: "Nombre total de webhooks envoyés aux clients callback",
			},
			[]string{"status"},
		),
		RateLimitHitsTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kyc_rate_limit_hits_total",
				Help: "Nombre total de requêtes bloquées par le rate limiter",
			},
			[]string{"endpoint", "limit_type"},
		),
		IdempotencyReplaysTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kyc_idempotency_replays_total",
				Help: "Nombre de réponses servies depuis le cache idempotency",
			},
			[]string{"endpoint"},
		),
		DocumentsDeletedTotal: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "kyc_documents_deleted_total",
				Help: "Nombre total de documents supprimés de MinIO par raison",
			},
			[]string{"reason"},
		),
		RequestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kyc_request_duration_seconds",
				Help:    "Durée des requêtes HTTP en secondes",
				Buckets: []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2, 5, 10},
			},
			[]string{"method", "endpoint", "status_code"},
		),
		ProviderDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kyc_provider_duration_seconds",
				Help:    "Durée des appels aux providers d'identité en secondes",
				Buckets: []float64{0.1, 0.5, 1, 2, 5, 10, 30},
			},
			[]string{"provider"},
		),
		DatabaseQueryDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "kyc_database_query_duration_seconds",
				Help:    "Durée des requêtes PostgreSQL en secondes",
				Buckets: []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.5},
			},
			[]string{"operation"},
		),
		VerificationsPending: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "kyc_verifications_pending",
				Help: "Nombre de vérifications en attente de traitement par pays",
			},
			[]string{"country_code"},
		),
		VerificationsProcessing: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "kyc_verifications_processing",
				Help: "Nombre de vérifications en cours de traitement par les providers",
			},
		),
		ProviderScore: prometheus.NewSummaryVec(
			prometheus.SummaryOpts{
				Name:       "kyc_provider_score",
				Help:       "Distribution des scores de confiance retournés par les providers",
				Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.95: 0.005, 0.99: 0.001},
			},
			[]string{"provider", "country_code"},
		),
		CleanupJobLastRun: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "kyc_cleanup_job_last_run_timestamp",
				Help: "Timestamp Unix du dernier run réussi de chaque tâche de cleanup",
			},
			[]string{"task"},
		),
		StorageDocumentsSizeBytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "kyc_storage_documents_size_bytes",
				Help: "Taille estimée des documents stockés dans MinIO en bytes",
			},
			[]string{"state"},
		),
	}

	collectors := []prometheus.Collector{
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
		m.VerificationsProcessing,
		m.ProviderScore,
		m.CleanupJobLastRun,
		m.StorageDocumentsSizeBytes,
	}
	for _, c := range collectors {
		if err := reg.Register(c); err != nil {
			panic("kycmetrics register: " + err.Error())
		}
	}
	return m
}

func NewDefaultMetrics() *KYCMetrics {
	return NewKYCMetrics(prometheus.DefaultRegisterer)
}

func normalizePath(path string) string {
	segments := []byte(path)
	start := -1
	for i, b := range segments {
		if b == '/' {
			if start >= 0 {
				seg := string(segments[start:i])
				if uuidRegex.MatchString(seg) {
					copy(segments[start:i], []byte(":id"))
				}
			}
			start = i + 1
		}
	}
	if start >= 0 {
		seg := string(segments[start:])
		if uuidRegex.MatchString(seg) {
			return string(segments[:start]) + ":id"
		}
	}
	return string(segments)
}

func PrometheusMiddleware(m *KYCMetrics) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()
		if err := c.Next(); err != nil {
			return err
		}
		duration := time.Since(start).Seconds()
		m.RequestDuration.WithLabelValues(
			c.Method(),
			normalizePath(c.Path()),
			strconv.Itoa(c.Response().StatusCode()),
		).Observe(duration)
		return nil
	}
}