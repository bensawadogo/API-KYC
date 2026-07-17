package observability

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestKYCInitiatedCounter(t *testing.T) {
	KYCInitiated.WithLabelValues("CI", "passport").Inc()
	KYCInitiated.WithLabelValues("SN", "nid").Inc()

	count := testutil.CollectAndCount(KYCInitiated)
	assert.GreaterOrEqual(t, count, 1)
}

func TestLabelsAreSet(t *testing.T) {
	KYCCompleted.WithLabelValues("approved", "smileid", "CI").Inc()
	KYCCompleted.WithLabelValues("rejected", "youverify", "SN").Inc()

	count := testutil.CollectAndCount(KYCCompleted)
	assert.GreaterOrEqual(t, count, 1)
}

func TestDLQSizeGauge(t *testing.T) {
	DLQSize.Set(5)
	DLQSize.Set(3)

	val := testutil.ToFloat64(DLQSize)
	assert.Equal(t, float64(3), val)
}

func init() {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
}
