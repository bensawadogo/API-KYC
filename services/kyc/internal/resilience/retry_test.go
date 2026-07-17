package resilience_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/datakeys/kyc-service/internal/resilience"
)

func TestDo_SuccessFirstAttempt(t *testing.T) {
	callCount := 0
	fn := func() error {
		callCount++
		return nil
	}

	err := resilience.Do(context.Background(), resilience.DefaultRetryConfig(), fn)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}

func TestDo_SuccessAfterRetry(t *testing.T) {
	callCount := 0
	fn := func() error {
		callCount++
		if callCount < 3 {
			return errors.New("temporary error")
		}
		return nil
	}

	cfg := resilience.DefaultRetryConfig()
	cfg.MaxAttempts = 5
	cfg.BaseDelay = time.Millisecond

	err := resilience.Do(context.Background(), cfg, fn)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestDo_ExhaustedRetries(t *testing.T) {
	callCount := 0
	fn := func() error {
		callCount++
		return errors.New("always fails")
	}

	cfg := resilience.DefaultRetryConfig()
	cfg.MaxAttempts = 3
	cfg.BaseDelay = time.Millisecond

	err := resilience.Do(context.Background(), cfg, fn)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestDo_PermanentError(t *testing.T) {
	callCount := 0
	fn := func() error {
		callCount++
		return resilience.Permanent(errors.New("validation error"))
	}

	cfg := resilience.DefaultRetryConfig()
	cfg.MaxAttempts = 5
	cfg.BaseDelay = time.Millisecond

	err := resilience.Do(context.Background(), cfg, fn)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if callCount != 1 {
		t.Errorf("expected 1 call (no retry for permanent), got %d", callCount)
	}
}

func TestDo_ContextCancelled(t *testing.T) {
	callCount := 0
	fn := func() error {
		callCount++
		return errors.New("temporary error")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	cfg := resilience.DefaultRetryConfig()
	cfg.MaxAttempts = 5
	cfg.BaseDelay = 100 * time.Millisecond

	err := resilience.Do(ctx, cfg, fn)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
	if callCount != 0 {
		t.Errorf("expected 0 calls (context already cancelled), got %d", callCount)
	}
}

func TestDo_ContextCancelledDuringRetry(t *testing.T) {
	callCount := 0
	ctx, cancel := context.WithCancel(context.Background())

	fn := func() error {
		callCount++
		if callCount == 1 {
			cancel()
		}
		return errors.New("temporary error")
	}

	cfg := resilience.DefaultRetryConfig()
	cfg.MaxAttempts = 5
	cfg.BaseDelay = time.Millisecond

	err := resilience.Do(ctx, cfg, fn)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestCalculateDelay_Bounds(t *testing.T) {
	cfg := resilience.DefaultRetryConfig()
	cfg.BaseDelay = 100 * time.Millisecond
	cfg.MaxDelay = 5 * time.Second

	for attempt := 0; attempt < 3; attempt++ {
		delays := make([]time.Duration, 0, 100)
		for i := 0; i < 100; i++ {
			delay := resilience.CalculateDelay(cfg, attempt)
			delays = append(delays, delay)
			minDelay := cfg.BaseDelay / 2
			if delay < minDelay {
				t.Errorf("attempt=%d iter=%d: delay %v < min %v",
					attempt, i, delay, minDelay)
			}
			if delay > cfg.MaxDelay {
				t.Errorf("attempt=%d iter=%d: delay %v > max %v",
					attempt, i, delay, cfg.MaxDelay)
			}
		}
		allSame := true
		for i := 1; i < len(delays); i++ {
			if delays[i] != delays[0] {
				allSame = false
				break
			}
		}
		if allSame {
			t.Errorf("attempt=%d: all delays identical, jitter not active", attempt)
		}
	}
}
