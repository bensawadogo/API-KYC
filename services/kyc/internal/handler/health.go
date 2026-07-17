package handler

import (
	"os"

	"github.com/datakeys/kyc-service/internal/resilience"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	db     *pgxpool.Pool
	redis  *redis.Client
	router *resilience.FallbackRouter
	dlq    resilience.DLQInterface
}

func NewHealthHandler(
	db *pgxpool.Pool,
	redis *redis.Client,
	router *resilience.FallbackRouter,
	dlq resilience.DLQInterface,
) *HealthHandler {
	return &HealthHandler{
		db:     db,
		redis:  redis,
		router: router,
		dlq:    dlq,
	}
}

func (h *HealthHandler) Live(c fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"service": "kyc-service",
	})
}

func (h *HealthHandler) Ready(c fiber.Ctx) error {
	checks := fiber.Map{}
	allOK := true

	if h.db != nil {
		if err := h.db.Ping(c.Context()); err != nil {
			checks["postgres"] = "down"
			allOK = false
		} else {
			checks["postgres"] = "up"
		}
	} else {
		checks["postgres"] = "not_configured"
	}

	if h.redis != nil {
		if err := h.redis.Ping(c.Context()).Err(); err != nil {
			checks["redis"] = "down"
		} else {
			checks["redis"] = "up"
		}
	} else {
		checks["redis"] = "not_configured"
	}

	if h.router != nil {
		checks["providers"] = h.router.ProviderStatuses()
	} else {
		checks["providers"] = "not_configured"
	}

	if h.dlq != nil {
		dlqLen, _ := h.dlq.Len(c.Context())
		checks["dlq_size"] = dlqLen
		if dlqLen > 100 {
			checks["dlq_alert"] = "HIGH — vérifier les providers"
		}
	} else {
		checks["dlq_size"] = 0
	}

	status := 200
	if !allOK {
		status = 503
	}

	version := os.Getenv("SERVICE_VERSION")
	if version == "" {
		version = "dev"
	}

	return c.Status(status).JSON(fiber.Map{
		"status":  map[bool]string{true: "ready", false: "degraded"}[allOK],
		"version": version,
		"checks":  checks,
	})
}
