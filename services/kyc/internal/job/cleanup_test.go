package job_test

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/datakeys/kyc-service/internal/job"
	"github.com/datakeys/kyc-service/internal/observability"
	dto "github.com/prometheus/client_model/go"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestCleanupJob_StartStop(t *testing.T) {
	logger := zap.NewNop()
	j := job.NewCleanupJob(nil, nil, nil, nil, 1825, logger)

	done := make(chan struct{})
	go func() {
		j.Start()
		close(done)
	}()

	time.Sleep(250 * time.Millisecond)

	stopDone := make(chan struct{})
	go func() {
		j.Stop()
		close(stopDone)
	}()

	select {
	case <-stopDone:
	case <-time.After(2 * time.Second):
		t.Fatal("Stop() blocked")
	}
	<-done
}

func TestRunNow_UpdatesDLQGauge(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer rdb.Close()

	for i := 0; i < 3; i++ {
		mr.Lpush("kyc:dlq", `{"verification_id":"test"}`)
	}

	logger := zap.NewNop()
	j := job.NewCleanupJob(nil, nil, nil, rdb, 1825, logger)

	observability.DLQSize.Set(0)
	j.RunNow()

	var metric dto.Metric
	if err := observability.DLQSize.Write(&metric); err != nil {
		t.Fatalf("DLQSize.Write failed: %v", err)
	}
	assert.Equal(t, float64(3), metric.GetGauge().GetValue())
}

func TestPurgeExpired_CallsDeleteDocument(t *testing.T) {
	t.Skip("PostgreSQL non disponible")
}
