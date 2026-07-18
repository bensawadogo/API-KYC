package compliance

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type PostgresConsentManager struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

func NewConsentManager(pool *pgxpool.Pool, logger *zap.Logger) *PostgresConsentManager {
	return &PostgresConsentManager{pool: pool, logger: logger}
}

func (m *PostgresConsentManager) RunMigration(ctx context.Context) error {
	sql := `
	CREATE TABLE IF NOT EXISTS kyc_consent_records (
		id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		phone_hash      VARCHAR(64) NOT NULL,
		event_type      VARCHAR(20) NOT NULL,
		consent_version VARCHAR(20) NOT NULL DEFAULT 'v1',
		ip_address      VARCHAR(45) DEFAULT '',
		created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
	);
	CREATE INDEX IF NOT EXISTS idx_consent_phone_hash ON kyc_consent_records(phone_hash);
	CREATE INDEX IF NOT EXISTS idx_consent_created_at ON kyc_consent_records(created_at);
	`
	_, err := m.pool.Exec(ctx, sql)
	return err
}

func hashPhone(phone string) string {
	h := sha256.Sum256([]byte(phone))
	return hex.EncodeToString(h[:])
}

func (m *PostgresConsentManager) RecordConsent(ctx context.Context, phone string, version string, ipAddress string) error {
	_, err := m.pool.Exec(ctx, `
		INSERT INTO kyc_consent_records (phone_hash, event_type, consent_version, ip_address)
		VALUES ($1, 'granted', $2, $3)
	`, hashPhone(phone), version, ipAddress)
	if err != nil {
		return fmt.Errorf("record consent: %w", err)
	}
	return nil
}

func (m *PostgresConsentManager) RecordWithdrawal(ctx context.Context, phone string, ip string) error {
	_, err := m.pool.Exec(ctx, `
		INSERT INTO kyc_consent_records (phone_hash, event_type, ip_address)
		VALUES ($1, 'withdrawn', $2)
	`, hashPhone(phone), ip)
	if err != nil {
		return fmt.Errorf("record withdrawal: %w", err)
	}
	return nil
}

func (m *PostgresConsentManager) HasActiveConsent(ctx context.Context, phone string) (bool, error) {
	var lastEvent string
	err := m.pool.QueryRow(ctx, `
		SELECT event_type FROM kyc_consent_records
		WHERE phone_hash = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, hashPhone(phone)).Scan(&lastEvent)
	if err != nil {
		return false, nil
	}
	return lastEvent == "granted", nil
}

func (m *PostgresConsentManager) GetConsentHistory(ctx context.Context, phone string) ([]ConsentRecord, error) {
	rows, err := m.pool.Query(ctx, `
		SELECT id, phone_hash, event_type, COALESCE(consent_version, 'v1'),
		       COALESCE(ip_address, ''), created_at
		FROM kyc_consent_records
		WHERE phone_hash = $1
		ORDER BY created_at DESC
	`, hashPhone(phone))
	if err != nil {
		return nil, fmt.Errorf("consent history: %w", err)
	}
	defer rows.Close()

	var records []ConsentRecord
	for rows.Next() {
		var r ConsentRecord
		if err := rows.Scan(&r.ID, &r.PhoneHash, &r.EventType, &r.ConsentVersion, &r.IPAddress, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan consent record: %w", err)
		}
		records = append(records, r)
	}
	return records, rows.Err()
}
