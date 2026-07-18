package resilience_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/datakeys/kyc-service/internal/resilience"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCircuitBreaker_ClosedByDefault(t *testing.T) {
	cb := resilience.NewCircuitBreaker(resilience.DefaultCBConfig("test"), zap.NewNop())
	assert.Equal(t, "closed", cb.State())
	assert.False(t, cb.IsOpen())
}

func TestCircuitBreaker_OpensAfterFailures(t *testing.T) {
	cfg := resilience.DefaultCBConfig("test")
	cfg.MaxFailures = 3
	cb := resilience.NewCircuitBreaker(cfg, zap.NewNop())
	fail := errors.New("provider error")
	for i := 0; i < 3; i++ {
		_, err := cb.Execute(func() (interface{}, error) {
			return nil, fail
		})
		if i < 2 {
			assert.Error(t, err)
		}
	}
	assert.True(t, cb.IsOpen())
	assert.Equal(t, "open", cb.State())
}

func TestCircuitBreaker_SuccessKeepsClosed(t *testing.T) {
	cb := resilience.NewCircuitBreaker(resilience.DefaultCBConfig("test"), zap.NewNop())
	for i := 0; i < 5; i++ {
		res, err := cb.Execute(func() (interface{}, error) {
			return "ok", nil
		})
		assert.NoError(t, err)
		assert.Equal(t, "ok", res)
	}
	assert.Equal(t, "closed", cb.State())
}

func TestCircuitBreaker_RejectsWhenOpen(t *testing.T) {
	cfg := resilience.DefaultCBConfig("test")
	cfg.MaxFailures = 1
	cb := resilience.NewCircuitBreaker(cfg, zap.NewNop())
	_, err := cb.Execute(func() (interface{}, error) {
		return nil, errors.New("fail")
	})
	assert.Error(t, err)

	assert.True(t, cb.IsOpen())
	_, err = cb.Execute(func() (interface{}, error) {
		return "should not run", nil
	})
	assert.Error(t, err)
}

func TestDLQ_EnqueueAndLen(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	dlq := resilience.NewDLQ(rdb, zap.NewNop())
	ctx := context.Background()

	err = dlq.Enqueue(ctx, "verification-uuid-001")
	assert.NoError(t, err)

	err = dlq.Enqueue(ctx, "verification-uuid-002")
	assert.NoError(t, err)

	size, err := dlq.Len(ctx)
	assert.NoError(t, err)
	assert.Equal(t, int64(2), size)
}

func TestDLQ_Dequeue(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	dlq := resilience.NewDLQ(rdb, zap.NewNop())
	ctx := context.Background()

	require.NoError(t, dlq.Enqueue(ctx, "uuid-aaa"))
	require.NoError(t, dlq.Enqueue(ctx, "uuid-bbb"))

	entries, err := dlq.Dequeue(ctx, 10)
	assert.NoError(t, err)
	assert.Len(t, entries, 2)

	ids := []string{entries[0].VerificationID, entries[1].VerificationID}
	assert.Contains(t, ids, "uuid-aaa")
	assert.Contains(t, ids, "uuid-bbb")
}
