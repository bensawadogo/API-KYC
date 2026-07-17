package audit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const migrationSQL = `
CREATE TABLE IF NOT EXISTS kyc_audit_log (
  id              BIGSERIAL PRIMARY KEY,
  event_type      VARCHAR(50)  NOT NULL,
  verification_id UUID         NOT NULL,
  phone_hash      VARCHAR(64)  NOT NULL,
  country_code    VARCHAR(2)   NOT NULL,
  doc_type        VARCHAR(30)  NOT NULL,
  provider        VARCHAR(50),
  status_before   VARCHAR(20),
  status_after    VARCHAR(20),
  score           DECIMAL(4,3),
  flags           TEXT[],
  ip_address      VARCHAR(45),
  user_agent      TEXT,
  duration_ms     INT,
  error_msg       TEXT,
  metadata        JSONB,
  created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_audit_verification_id ON kyc_audit_log(verification_id);
CREATE INDEX IF NOT EXISTS idx_audit_phone_hash ON kyc_audit_log(phone_hash);
CREATE INDEX IF NOT EXISTS idx_audit_created_at ON kyc_audit_log(created_at);`

const (
	EventInitiated    = "kyc.initiated"
	EventProcessing   = "kyc.processing"
	EventApproved     = "kyc.approved"
	EventRejected     = "kyc.rejected"
	EventManualReview = "kyc.manual_review"
	EventExpired      = "kyc.expired"
	EventWebhookSent  = "kyc.webhook_sent"
	EventWebhookFail  = "kyc.webhook_failed"
	EventProviderFail = "kyc.provider_failed"
	EventDocDeleted   = "kyc.document_deleted"
)

type AuditEntry struct {
	EventType      string
	VerificationID string
	Phone          string
	CountryCode    string
	DocType        string
	Provider       string
	StatusBefore   string
	StatusAfter    string
	Score          float64
	Flags          []string
	IPAddress      string
	UserAgent      string
	DurationMS     int
	ErrorMsg       string
	Metadata       map[string]interface{}
}

type AuditLogger struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewAuditLogger(pool *pgxpool.Pool, logger *zap.Logger) (*AuditLogger, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := pool.Exec(ctx, migrationSQL); err != nil {
		return nil, fmt.Errorf("audit migration: %w", err)
	}

	logger.Info("audit_logger initialisé")
	return &AuditLogger{pool: pool, logger: logger}, nil
}

func (a *AuditLogger) RunMigration(ctx context.Context) error {
	_, err := a.pool.Exec(ctx, migrationSQL)
	return err
}

func hashPhone(phone string) string {
	h := sha256.Sum256([]byte(phone))
	return hex.EncodeToString(h[:])
}

func (a *AuditLogger) Log(ctx context.Context, entry AuditEntry) {
	if a.pool == nil {
		a.logger.Error("audit log skipped: pool is nil")
		return
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	phoneHash := hashPhone(entry.Phone)

	var meta []byte
	if entry.Metadata != nil {
		meta, _ = json.Marshal(entry.Metadata)
	}

	_, err := a.pool.Exec(ctx, `
		INSERT INTO kyc_audit_log (
			event_type, verification_id, phone_hash, country_code, doc_type,
			provider, status_before, status_after, score, flags,
			ip_address, user_agent, duration_ms, error_msg, metadata
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		entry.EventType, entry.VerificationID, phoneHash, entry.CountryCode, entry.DocType,
		entry.Provider, entry.StatusBefore, entry.StatusAfter, entry.Score, entry.Flags,
		entry.IPAddress, entry.UserAgent, entry.DurationMS, entry.ErrorMsg, meta,
	)
	if err != nil {
		a.logger.Error("audit log insert échoué",
			zap.String("event_type", entry.EventType),
			zap.String("verification_id", entry.VerificationID),
			zap.Error(err),
		)
	}
}

func (a *AuditLogger) LogAsync(ctx context.Context, entry AuditEntry) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		a.Log(ctx, entry)
	}()
	_ = ctx
}

func (a *AuditLogger) GetByVerificationID(ctx context.Context, verificationID string) ([]AuditEntry, error) {
	rows, err := a.pool.Query(ctx, `
		SELECT event_type, verification_id, phone_hash, country_code, doc_type,
		       COALESCE(provider,''), COALESCE(status_before,''), COALESCE(status_after,''),
		       COALESCE(score,0), COALESCE(flags,'{}'), COALESCE(ip_address,''),
		       COALESCE(user_agent,''), COALESCE(duration_ms,0), COALESCE(error_msg,''),
		       metadata
		FROM kyc_audit_log
		WHERE verification_id = $1
		ORDER BY created_at ASC
	`, verificationID)
	if err != nil {
		return nil, fmt.Errorf("audit query: %w", err)
	}
	defer rows.Close()

	var entries []AuditEntry
	for rows.Next() {
		var e AuditEntry
		var phoneHash, metaStr string
		if err := rows.Scan(
			&e.EventType, &e.VerificationID, &phoneHash, &e.CountryCode, &e.DocType,
			&e.Provider, &e.StatusBefore, &e.StatusAfter,
			&e.Score, &e.Flags, &e.IPAddress,
			&e.UserAgent, &e.DurationMS, &e.ErrorMsg, &metaStr,
		); err != nil {
			return nil, fmt.Errorf("audit scan: %w", err)
		}
		if metaStr != "" {
			_ = json.Unmarshal([]byte(metaStr), &e.Metadata)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}