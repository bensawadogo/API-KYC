package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	KYCInitiated = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "kyc",
			Name:      "initiated_total",
			Help:      "Nombre total de vérifications initiées",
		},
		[]string{"country_code", "doc_type"},
	)

	KYCCompleted = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "kyc",
			Name:      "completed_total",
			Help:      "Vérifications terminées par statut final",
		},
		[]string{"status", "provider", "country_code"},
	)

	KYCAMLFlagged = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "kyc",
			Name:      "aml_flagged_total",
			Help:      "Vérifications avec flag AML/sanctions",
		},
		[]string{"flag_type", "country_code"},
	)

	KYCDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "kyc",
			Name:      "process_duration_seconds",
			Help:      "Durée du traitement KYC de bout en bout",
			Buckets:   []float64{0.5, 1, 2, 5, 10, 30, 60},
		},
		[]string{"provider", "country_code"},
	)

	ProviderRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "kyc",
			Subsystem: "provider",
			Name:      "requests_total",
			Help:      "Appels vers les providers par résultat",
		},
		[]string{"provider", "result"},
	)

	ProviderDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "kyc",
			Subsystem: "provider",
			Name:      "duration_seconds",
			Help:      "Latence des appels providers",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"provider"},
	)

	ProviderCircuitState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "kyc",
			Subsystem: "provider",
			Name:      "circuit_state",
			Help:      "État circuit breaker (0=closed, 1=half-open, 2=open)",
		},
		[]string{"provider"},
	)

	DLQSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "kyc",
			Name:      "dlq_size",
			Help:      "Nombre d'entrées dans la Dead Letter Queue",
		},
	)

	HTTPRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "kyc",
			Subsystem: "http",
			Name:      "requests_total",
			Help:      "Requêtes HTTP par route et statut",
		},
		[]string{"method", "path", "status"},
	)

	HTTPDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "kyc",
			Subsystem: "http",
			Name:      "request_duration_seconds",
			Help:      "Latence HTTP par route",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.3, 0.5, 1, 2},
		},
		[]string{"method", "path"},
	)

	RateLimitHits = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "kyc",
			Subsystem: "http",
			Name:      "rate_limit_hits_total",
			Help:      "Requêtes rejetées par rate limiting",
		},
		[]string{"client_name"},
	)
)
