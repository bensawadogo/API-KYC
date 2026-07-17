package resilience

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type DLQEntry struct {
	VerificationID string    `json:"verification_id"`
	EnqueuedAt     time.Time `json:"enqueued_at"`
	Attempts       int       `json:"attempts"`
	LastError      string    `json:"last_error"`
}

type DLQ struct {
	redis  *redis.Client
	key    string
	maxLen int64
	ttl    time.Duration
	logger *zap.Logger
}

var _ DLQInterface = (*DLQ)(nil)

func NewDLQ(redis *redis.Client, logger *zap.Logger) *DLQ {
	return &DLQ{
		redis:  redis,
		key:    "kyc:dlq",
		maxLen: 1000,
		ttl:    7 * 24 * time.Hour,
		logger: logger,
	}
}

func (d *DLQ) Enqueue(ctx context.Context, verificationID string) error {
	entry := DLQEntry{
		VerificationID: verificationID,
		EnqueuedAt:     time.Now(),
		Attempts:       1,
	}
	data, _ := json.Marshal(entry)

	pipe := d.redis.Pipeline()
	pipe.LPush(ctx, d.key, string(data))
	pipe.LTrim(ctx, d.key, 0, d.maxLen-1)
	_, err := pipe.Exec(ctx)

	if err != nil {
		d.logger.Error("DLQ enqueue failed",
			zap.String("verification_id", verificationID),
			zap.Error(err),
		)
	} else {
		d.logger.Warn("verification added to DLQ",
			zap.String("verification_id", verificationID))
	}
	return err
}

func (d *DLQ) Dequeue(ctx context.Context, count int) ([]DLQEntry, error) {
	results, err := d.redis.LRange(ctx, d.key, 0, int64(count-1)).Result()
	if err != nil {
		return nil, err
	}
	entries := make([]DLQEntry, 0, len(results))
	for _, r := range results {
		var e DLQEntry
		if json.Unmarshal([]byte(r), &e) == nil {
			entries = append(entries, e)
		}
	}
	return entries, nil
}

func (d *DLQ) Len(ctx context.Context) (int64, error) {
	return d.redis.LLen(ctx, d.key).Result()
}

type DLQInterface interface {
	Enqueue(ctx context.Context, verificationID string) error
	Len(ctx context.Context) (int64, error)
}
