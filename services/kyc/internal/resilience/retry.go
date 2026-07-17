package resilience

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"time"
)

type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
}

func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		BaseDelay:   500 * time.Millisecond,
		MaxDelay:    10 * time.Second,
		Multiplier:  2.0,
	}
}

func Do(ctx context.Context, cfg RetryConfig, fn func() error) error {
	var lastErr error
	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		lastErr = fn()
		if lastErr == nil {
			return nil
		}

		if IsPermanent(lastErr) {
			return lastErr
		}

		if attempt < cfg.MaxAttempts-1 {
			delay := CalculateDelay(cfg, attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}
	return fmt.Errorf("après %d tentatives: %w", cfg.MaxAttempts, lastErr)
}

func CalculateDelay(cfg RetryConfig, attempt int) time.Duration {
	base := float64(cfg.BaseDelay)
	exp := math.Pow(cfg.Multiplier, float64(attempt))
	delay := time.Duration(base * exp)
	if delay > cfg.MaxDelay {
		delay = cfg.MaxDelay
	}
	jitter := time.Duration(float64(delay/2) + rand.Float64()*float64(delay/2))
	return jitter
}

type PermanentError struct{ Err error }

func (e *PermanentError) Error() string { return e.Err.Error() }
func (e *PermanentError) Unwrap() error { return e.Err }

func Permanent(err error) error {
	return &PermanentError{Err: err}
}

func IsPermanent(err error) bool {
	var pe *PermanentError
	return errors.As(err, &pe)
}
