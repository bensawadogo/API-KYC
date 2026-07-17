package middleware

import (
	"regexp"
	"strconv"
	"time"

	"github.com/datakeys/kyc-service/internal/observability"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

func MetricsAndTracing() fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			c.Set("X-Request-ID", requestID)
		}

	tracer := observability.Tracer()
	spanName := c.Method() + " " + c.Path()
	ctx, span := tracer.Start(c.Context(), spanName,
		trace.WithSpanKind(trace.SpanKindServer),
		trace.WithAttributes(
			attribute.String("http.method", c.Method()),
			attribute.String("http.path", c.Path()),
			attribute.String("request.id", requestID),
		),
	)
	defer span.End()
	c.SetContext(ctx)

		err := c.Next()

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Response().StatusCode())
		path := normalizePath(c.Path())

		observability.HTTPRequests.WithLabelValues(c.Method(), path, status).Inc()
		observability.HTTPDuration.WithLabelValues(c.Method(), path).Observe(duration)

		span.SetAttributes(attribute.String("http.status_code", status))
		if c.Response().StatusCode() >= 500 {
			span.SetStatus(codes.Error, "server error")
		}

		return err
	}
}

var uuidRegex = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
var countryRegex = regexp.MustCompile(`/countries/[A-Z]{2}$`)

func normalizePath(path string) string {
	path = uuidRegex.ReplaceAllString(path, ":id")
	path = countryRegex.ReplaceAllString(path, "/countries/:code")
	return path
}
