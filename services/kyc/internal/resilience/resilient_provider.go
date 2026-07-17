package resilience

import (
	"context"
	"errors"
	"time"

	"github.com/datakeys/kyc-service/internal"
	"github.com/datakeys/kyc-service/internal/observability"
	"github.com/sony/gobreaker"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type ResilientProvider struct {
	provider internal.IdentityProvider
	cb       *CircuitBreaker
	retry    RetryConfig
	logger   *zap.Logger
}

func NewResilientProvider(
	p internal.IdentityProvider,
	cbCfg CBConfig,
	retryCfg RetryConfig,
	logger *zap.Logger,
) *ResilientProvider {
	return &ResilientProvider{
		provider: p,
		cb:       NewCircuitBreaker(cbCfg, logger),
		retry:    retryCfg,
		logger:   logger,
	}
}

func (r *ResilientProvider) Verify(ctx context.Context, req internal.ProviderRequest) (*internal.ProviderResult, error) {
	ctx, span := observability.Tracer().Start(ctx, r.provider.Name()+".Verify",
		trace.WithAttributes(
			attribute.String("provider", r.provider.Name()),
			attribute.String("verification_id", req.VerificationID),
			attribute.String("country_code", req.CountryCode),
		),
	)
	defer span.End()

	start := time.Now()
	var result *internal.ProviderResult

	err := Do(ctx, r.retry, func() error {
		res, err := r.cb.Execute(func() (interface{}, error) {
			return r.provider.Verify(ctx, req)
		})
		if err != nil {
			if errors.Is(err, gobreaker.ErrOpenState) {
				return Permanent(err)
			}
			return err
		}
		result = res.(*internal.ProviderResult)
		return nil
	})

	duration := time.Since(start).Seconds()
	state := r.cb.State()
	resultLabel := "success"
	if err != nil {
		resultLabel = "failure"
	}
	observability.ProviderRequests.WithLabelValues(r.provider.Name(), resultLabel).Inc()
	observability.ProviderDuration.WithLabelValues(r.provider.Name()).Observe(duration)
	observability.ProviderCircuitState.WithLabelValues(r.provider.Name()).Set(float64(circuitStateToInt(state)))

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())

		r.logger.Error("provider failed after retries",
			zap.String("provider", r.provider.Name()),
			zap.String("verification_id", req.VerificationID),
			zap.String("circuit_state", state),
			zap.Error(err),
		)
	}
	return result, err
}

func circuitStateToInt(state string) int {
	switch state {
	case "closed":
		return 0
	case "half-open":
		return 1
	case "open":
		return 2
	default:
		return 0
	}
}

func (r *ResilientProvider) SupportedCountries() []string {
	return r.provider.SupportedCountries()
}

func (r *ResilientProvider) Name() string {
	return r.provider.Name()
}

func (r *ResilientProvider) IsAvailable() bool {
	return !r.cb.IsOpen()
}
