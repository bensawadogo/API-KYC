package middleware

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RateLimitConfig struct {
	RequestsPerMinute int
	RequestsPerHour   int
	RequestsPerDay    int
	KeyPrefix         string
}

type RateLimiter struct {
	rdb    *redis.Client
	logger *zap.Logger
}

func NewRateLimiter(rdb *redis.Client, logger *zap.Logger) *RateLimiter {
	return &RateLimiter{rdb: rdb, logger: logger}
}

// New creates a fiber.Handler with the given rate limit configuration.
func (rl *RateLimiter) New(cfg RateLimitConfig) fiber.Handler {
	if cfg.RequestsPerMinute <= 0 {
		cfg.RequestsPerMinute = 60
	}
	if cfg.RequestsPerHour <= 0 {
		cfg.RequestsPerHour = 200
	}
	if cfg.RequestsPerDay <= 0 {
		cfg.RequestsPerDay = 1000
	}
	if cfg.KeyPrefix == "" {
		cfg.KeyPrefix = "rl:kyc"
	}

	return func(c fiber.Ctx) error {
		ip := c.IP()
		apiKey := c.Get("X-API-Key")
		hash := sha256Hex(ip + ":" + apiKey)

		limits := []struct {
			suffix string
			ttl    int64
			limit  int
			name   string
		}{
			{":min", 60, cfg.RequestsPerMinute, "minute"},
			{":hour", 3600, cfg.RequestsPerHour, "hour"},
			{":day", 86400, cfg.RequestsPerDay, "day"},
		}

		for _, l := range limits {
			key := cfg.KeyPrefix + ":" + hash + l.suffix
			val, err := rl.rdb.Incr(c.Context(), key).Result()
			if err != nil {
				rl.logger.Warn("redis down, rate limit bypassé",
					zap.String("ip", ip),
					zap.Error(err),
				)
				return c.Next()
			}
			if val == 1 {
				rl.rdb.Expire(c.Context(), key, time.Duration(l.ttl)*time.Second)
			}
			if val > int64(l.limit) {
				ttl, _ := rl.rdb.TTL(c.Context(), key).Result()
				retryAfter := int(ttl.Seconds())
				if retryAfter < 1 {
					retryAfter = 1
				}
				c.Response().Header.Set("Retry-After", strconv.Itoa(retryAfter))
				return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
					"success":     false,
					"error":       "Rate limit dépassé",
					"retry_after": retryAfter,
					"limit_type":  l.name,
					"timestamp":   time.Now().UTC().Format(time.RFC3339),
				})
			}
		}

		now := time.Now().Unix()
		c.Response().Header.Set("X-RateLimit-Limit", strconv.Itoa(cfg.RequestsPerMinute))
		c.Response().Header.Set("X-RateLimit-Reset", strconv.FormatInt(now+60, 10))
		return c.Next()
	}
}

// Strict returns a handler limited to 10 req/min (POST /initiate).
func (rl *RateLimiter) Strict() fiber.Handler {
	return rl.New(RateLimitConfig{
		RequestsPerMinute: 10,
		RequestsPerHour:   50,
		RequestsPerDay:    200,
		KeyPrefix:         "rl:kyc",
	})
}

// Normal returns a handler limited to 60 req/min (GET /status).
func (rl *RateLimiter) Normal() fiber.Handler {
	return rl.New(RateLimitConfig{
		RequestsPerMinute: 60,
		RequestsPerHour:   200,
		RequestsPerDay:    1000,
		KeyPrefix:         "rl:kyc",
	})
}

// Relaxed returns a handler limited to 120 req/min (GET /countries).
func (rl *RateLimiter) Relaxed() fiber.Handler {
	return rl.New(RateLimitConfig{
		RequestsPerMinute: 120,
		RequestsPerHour:   500,
		RequestsPerDay:    2000,
		KeyPrefix:         "rl:kyc",
	})
}

// Webhook returns a handler limited to 100 req/min (provider callbacks).
func (rl *RateLimiter) Webhook() fiber.Handler {
	return rl.New(RateLimitConfig{
		RequestsPerMinute: 100,
		RequestsPerHour:   1000,
		RequestsPerDay:    5000,
		KeyPrefix:         "rl:kyc:wh",
	})
}

func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func RateLimitPerKey(redisClient *redis.Client) fiber.Handler {
	return func(c fiber.Ctx) error {
		if redisClient == nil {
			return c.Next()
		}
		keyID, ok := c.Locals("api_key_id").(string)
		if !ok || keyID == "" {
			return c.Next()
		}

		limit, _ := c.Locals("rate_limit").(int)
		if limit <= 0 {
			limit = 60
		}

		redisKey := fmt.Sprintf("rl:%s:%d", keyID, time.Now().Unix()/60)

		count, err := redisClient.Incr(c.Context(), redisKey).Result()
		if err != nil {
			return c.Next()
		}
		if count == 1 {
			redisClient.Expire(c.Context(), redisKey, 70*time.Second)
		}
		if count > int64(limit) {
			c.Response().Header.Set("X-RateLimit-Limit", strconv.Itoa(limit))
			c.Response().Header.Set("X-RateLimit-Remaining", "0")
			c.Response().Header.Set("Retry-After", "60")
			return c.Status(fiber.StatusTooManyRequests).JSON(AuthErrorResponse{
				Success: false,
				Error:   "KYC_RATE_001: Rate limit dépassé",
			})
		}

		c.Response().Header.Set("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Response().Header.Set("X-RateLimit-Remaining", strconv.Itoa(limit-int(count)))
		return c.Next()
	}
}