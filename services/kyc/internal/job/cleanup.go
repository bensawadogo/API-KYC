package job

import (
	"context"
	"sync"
	"time"

	"github.com/datakeys/kyc-service/internal"
	"github.com/datakeys/kyc-service/internal/observability"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type CleanupJob struct {
	repo      internal.KYCRepository
	storage   internal.DocumentStorage
	db        *pgxpool.Pool
	redis     *redis.Client
	logger    *zap.Logger
	interval  time.Duration
	retention int
	stopCh    chan struct{}
	wg        sync.WaitGroup
}

func NewCleanupJob(
	repo internal.KYCRepository,
	storage internal.DocumentStorage,
	db *pgxpool.Pool,
	redis *redis.Client,
	retention int,
	logger *zap.Logger,
) *CleanupJob {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &CleanupJob{
		repo:      repo,
		storage:   storage,
		db:        db,
		redis:     redis,
		logger:    logger,
		interval:  1 * time.Hour,
		retention: retention,
		stopCh:    make(chan struct{}),
	}
}

func (j *CleanupJob) Start() {
	j.wg.Add(1)
	go func() {
		defer j.wg.Done()

		j.RunNow()

		ticker := time.NewTicker(j.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				j.RunNow()
			case <-j.stopCh:
				return
			}
		}
	}()
	j.logger.Info("cleanup job started", zap.Duration("interval", j.interval))
}

func (j *CleanupJob) Stop() {
	close(j.stopCh)
	j.wg.Wait()
	j.logger.Info("cleanup job stopped")
}

func (j *CleanupJob) RunNow() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	j.purgeExpiredVerifications(ctx)
	j.purgeOrphanDocuments(ctx)
	j.updateDLQGauge(ctx)
}

func (j *CleanupJob) purgeExpiredVerifications(ctx context.Context) {
	if j.db == nil {
		return
	}

	rows, err := j.db.Query(ctx, `
		DELETE FROM kyc_verifications
		WHERE status IN ('pending', 'processing')
		AND expires_at < NOW()
		RETURNING id
	`)
	if err != nil {
		j.logger.Error("failed to purge expired verifications", zap.Error(err))
		return
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			j.logger.Error("scan expired id", zap.Error(err))
			continue
		}
		ids = append(ids, id)

		if err := j.storage.DeleteDocument(ctx, id); err != nil {
			j.logger.Warn("failed to delete document for expired verification",
				zap.String("id", id),
				zap.Error(err),
			)
		}
		j.logger.Info("expired verification purged", zap.String("id", id))
	}
	if err := rows.Err(); err != nil {
		j.logger.Error("rows iteration error", zap.Error(err))
	}

	if len(ids) > 0 {
		j.logger.Info("purged expired verifications", zap.Int("count", len(ids)))
	}
}

func (j *CleanupJob) purgeOrphanDocuments(ctx context.Context) {
	if j.db == nil || j.storage == nil {
		return
	}

	var tableExists bool
	err := j.db.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_name = 'kyc_audit_log'
		)
	`).Scan(&tableExists)
	if err != nil {
		j.logger.Error("check audit_log table", zap.Error(err))
		return
	}

	query := `
		SELECT id FROM kyc_verifications
		WHERE status = 'approved'
		AND processed_at < NOW() - INTERVAL '24 hours'
		LIMIT 100
	`
	if tableExists {
		query = `
			SELECT id FROM kyc_verifications
			WHERE status = 'approved'
			AND processed_at < NOW() - INTERVAL '24 hours'
			AND id NOT IN (
				SELECT DISTINCT verification_id
				FROM kyc_audit_log
				WHERE event_type = 'kyc.document_deleted'
			)
			LIMIT 100
		`
	} else {
		j.logger.Warn("kyc_audit_log table not found, skipping orphan purge")
	}

	rows, err := j.db.Query(ctx, query)
	if err != nil {
		j.logger.Error("failed to query orphan documents", zap.Error(err))
		return
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			j.logger.Error("scan orphan id", zap.Error(err))
			continue
		}
		ids = append(ids, id)

		if err := j.storage.DeleteDocument(ctx, id); err != nil {
			j.logger.Warn("failed to delete orphan document",
				zap.String("id", id),
				zap.Error(err),
			)
		}
		j.logger.Info("orphan document deleted", zap.String("id", id))
	}
	if err := rows.Err(); err != nil {
		j.logger.Error("rows iteration error", zap.Error(err))
	}

	if len(ids) > 0 {
		j.logger.Info("orphan documents deleted", zap.Int("count", len(ids)))
	}
}

func (j *CleanupJob) updateDLQGauge(ctx context.Context) {
	if j.redis == nil {
		return
	}

	count, err := j.redis.LLen(ctx, "kyc:dlq").Result()
	if err != nil {
		j.logger.Debug("failed to get DLQ size", zap.Error(err))
		return
	}
	observability.DLQSize.Set(float64(count))
}
