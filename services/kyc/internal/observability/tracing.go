package observability

import (
	"context"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type TracingConfig struct {
	Enabled     bool
	ServiceName string
	Environment string
}

func InitTracing(cfg TracingConfig) (func(), error) {
	if !cfg.Enabled {
		otel.SetTracerProvider(trace.NewNoopTracerProvider())
		return func() {}, nil
	}

	if cfg.ServiceName == "" {
		cfg.ServiceName = "kyc-service"
	}
	if cfg.Environment == "" {
		cfg.Environment = "development"
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		),
	)

	shutdown := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = tp.Shutdown(ctx)
	}

	return shutdown, nil
}

func Tracer() trace.Tracer {
	return otel.Tracer("kyc-service")
}
