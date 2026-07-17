package middleware

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type CachedResponse struct {
	StatusCode int    `json:"status_code"`
	Body       string `json:"body"`
}

type IdempotencyStore struct {
	redis  *redis.Client
	logger *zap.Logger
}

func NewIdempotencyStore(r *redis.Client, l *zap.Logger) *IdempotencyStore {
	return &IdempotencyStore{redis: r, logger: l}
}

func (s *IdempotencyStore) Middleware(c fiber.Ctx) error {
	key := c.Get("Idempotency-Key")
	if key == "" {
		return c.Next()
	}

	clientID, _ := c.Locals("api_key_id").(string)
	redisKey := fmt.Sprintf("idmp:%s:%s", clientID, key)

	lockPayload := `{"status":"processing"}`
	locked, err := s.redis.SetNX(c.Context(), redisKey, lockPayload, 30*time.Second).Result()
	if err != nil {
		s.logger.Warn("idempotency redis error, failing open",
			zap.String("key", key), zap.Error(err))
		return c.Next()
	}

	if !locked {
		existing, err := s.redis.Get(c.Context(), redisKey).Result()
		if err == nil {
			var cached CachedResponse
			if json.Unmarshal([]byte(existing), &cached) == nil && cached.StatusCode >= 200 {
				c.Response().Header.Set("Idempotency-Replayed", "true")
				c.Response().Header.Set("Idempotency-Key", key)
				return c.Status(cached.StatusCode).Send([]byte(cached.Body))
			}
		}
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"success":   false,
			"error":     "KYC_IDMP_001: Requête identique en cours de traitement",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}

	c.Response().Header.Set("Idempotency-Key", key)
	c.Response().Header.Set("Idempotent-Replayed", "false")

	err = c.Next()

	statusCode := c.Response().StatusCode()
	if statusCode >= 200 && statusCode < 300 {
		cachedResp := CachedResponse{
			StatusCode: statusCode,
			Body:       string(c.Response().Body()),
		}
		data, _ := json.Marshal(cachedResp)
		if setErr := s.redis.Set(c.Context(), redisKey, string(data), 24*time.Hour).Err(); setErr != nil {
			s.logger.Warn("idempotency cache set failed", zap.Error(setErr))
		}
	} else {
		s.redis.Del(c.Context(), redisKey)
	}

	return err
}
