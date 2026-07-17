package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/datakeys/kyc-service/internal"
	"github.com/datakeys/kyc-service/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(ctx context.Context, postgresURL string) (*PostgresRepository, error) {
	pool, err := pgxpool.New(ctx, postgresURL)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &PostgresRepository{pool: pool}, nil
}

func RunMigrations(pool *pgxpool.Pool) error {
	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()
	goose.SetBaseFS(nil)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("goose set dialect: %w", err)
	}
	if err := goose.Up(db, "./migrations"); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Close() {
	r.pool.Close()
}

func (r *PostgresRepository) Pool() *pgxpool.Pool {
	return r.pool
}

func (r *PostgresRepository) CreateVerification(ctx context.Context, v *model.VerificationResult) error {
	var expiresAt time.Time
	if v.ExpiresAt != "" {
		parsed, err := time.Parse(time.RFC3339, v.ExpiresAt)
		if err != nil {
			return fmt.Errorf("parse expires_at: %w", err)
		}
		expiresAt = parsed
	} else {
		expiresAt = time.Now().UTC().Add(time.Hour)
	}

	_, err := r.pool.Exec(ctx, `
		INSERT INTO kyc_verifications (
			id, phone, country_code, doc_type, status, score, provider, flags,
			callback_url, consent, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`,
		v.VerificationID,
		v.Phone,
		v.CountryCode,
		v.DocType,
		v.Status,
		v.Score,
		v.Provider,
		v.Flags,
		v.CallbackURL,
		v.Consent,
		expiresAt,
	)
	if err != nil {
		return fmt.Errorf("insert verification: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetVerification(ctx context.Context, id string) (*model.VerificationResult, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, phone, country_code, doc_type, status, COALESCE(score, 0),
		       COALESCE(provider, ''), COALESCE(flags, '{}'), COALESCE(callback_url, ''),
		       consent, processed_at, expires_at
		FROM kyc_verifications
		WHERE id = $1
	`, id)

	return scanVerification(row)
}

func (r *PostgresRepository) UpdateStatus(ctx context.Context, id, status string, score float64, flags []string, provider string) error {
	_, err := r.pool.Exec(ctx, `
		UPDATE kyc_verifications
		SET status = $2, score = $3, flags = $4, provider = $5,
		    processed_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`, id, status, score, flags, provider)
	if err != nil {
		return fmt.Errorf("update verification status: %w", err)
	}
	return nil
}

func (r *PostgresRepository) ListByPhone(ctx context.Context, phone string, limit int) ([]*model.VerificationResult, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := r.pool.Query(ctx, `
		SELECT id, phone, country_code, doc_type, status, COALESCE(score, 0),
		       COALESCE(provider, ''), COALESCE(flags, '{}'), COALESCE(callback_url, ''),
		       consent, processed_at, expires_at
		FROM kyc_verifications
		WHERE phone = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, phone, limit)
	if err != nil {
		return nil, fmt.Errorf("list verifications by phone: %w", err)
	}
	defer rows.Close()

	results := make([]*model.VerificationResult, 0)
	for rows.Next() {
		v, err := scanVerification(rows)
		if err != nil {
			return nil, err
		}
		results = append(results, v)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate verifications: %w", err)
	}
	return results, nil
}

func (r *PostgresRepository) ExistsApproved(ctx context.Context, phone, countryCode string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM kyc_verifications
			WHERE phone = $1 AND country_code = $2
			  AND status = 'approved' AND expires_at > NOW()
		)
	`, phone, countryCode).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("check approved verification: %w", err)
	}
	return exists, nil
}

func (r *PostgresRepository) SaveAMLResult(ctx context.Context, verificationID string, result *internal.AMLResult) error {
	raw, _ := json.Marshal(result.Matches)
	_, err := r.pool.Exec(ctx, `
		INSERT INTO kyc_aml_results
		(verification_id, is_sanctioned, is_pep, aml_score, matches_count,
		 source, screened_at, raw_matches)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`,
		verificationID,
		result.IsSanctioned,
		result.IsPEP,
		result.Score,
		len(result.Matches),
		result.Source,
		result.ScreenedAt,
		raw,
	)
	if err != nil {
		return fmt.Errorf("insert aml result: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetAMLResult(ctx context.Context, verificationID string) (*internal.AMLResult, error) {
	var (
		result       internal.AMLResult
		matchesCount int
		raw          []byte
	)
	err := r.pool.QueryRow(ctx, `
		SELECT is_sanctioned, is_pep, COALESCE(aml_score, 0),
		       COALESCE(matches_count, 0), COALESCE(source, ''),
		       screened_at, COALESCE(raw_matches, '[]'::jsonb)
		FROM kyc_aml_results
		WHERE verification_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, verificationID).Scan(
		&result.IsSanctioned,
		&result.IsPEP,
		&result.Score,
		&matchesCount,
		&result.Source,
		&result.ScreenedAt,
		&raw,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("aml result not found")
		}
		return nil, fmt.Errorf("get aml result: %w", err)
	}
	if len(raw) > 0 {
		json.Unmarshal(raw, &result.Matches)
	}
	return &result, nil
}

func (r *PostgresRepository) FindByPrefix(ctx context.Context, prefix string) (*model.APIKey, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, client_name, key_hash, key_prefix, scopes, rate_limit,
		       is_active, last_used_at, expires_at, created_at
		FROM api_keys
		WHERE key_prefix = $1 AND is_active = true
		  AND (expires_at IS NULL OR expires_at > NOW())
	`, prefix)
	return scanAPIKey(row)
}

func (r *PostgresRepository) ValidateKey(ctx context.Context, rawKey string) (*model.APIKey, error) {
	if len(rawKey) < 12 {
		return nil, fmt.Errorf("invalid key length")
	}
	prefix := rawKey[:12]
	if !strings.HasPrefix(prefix, "dk_") {
		return nil, fmt.Errorf("invalid key prefix")
	}

	apiKey, err := r.FindByPrefix(ctx, prefix)
	if err != nil {
		return nil, fmt.Errorf("find by prefix: %w", err)
	}

	hash := sha256.Sum256([]byte(rawKey))
	if hex.EncodeToString(hash[:]) != apiKey.KeyHash {
		return nil, fmt.Errorf("key hash mismatch")
	}

	go func() {
		_ = r.UpdateLastUsed(context.Background(), apiKey.ID)
	}()

	return apiKey, nil
}

func (r *PostgresRepository) UpdateLastUsed(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, `UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("update last used: %w", err)
	}
	return nil
}

type scannable interface {
	Scan(dest ...any) error
}

func scanVerification(row scannable) (*model.VerificationResult, error) {
	var v model.VerificationResult
	var processedAt *time.Time
	var expiresAt time.Time

	err := row.Scan(
		&v.VerificationID,
		&v.Phone,
		&v.CountryCode,
		&v.DocType,
		&v.Status,
		&v.Score,
		&v.Provider,
		&v.Flags,
		&v.CallbackURL,
		&v.Consent,
		&processedAt,
		&expiresAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("verification not found")
		}
		return nil, fmt.Errorf("scan verification: %w", err)
	}

	if processedAt != nil {
		v.ProcessedAt = processedAt.UTC().Format(time.RFC3339)
	}
	v.ExpiresAt = expiresAt.UTC().Format(time.RFC3339)
	return &v, nil
}

func scanAPIKey(row scannable) (*model.APIKey, error) {
	var k model.APIKey
	err := row.Scan(
		&k.ID,
		&k.ClientName,
		&k.KeyHash,
		&k.KeyPrefix,
		&k.Scopes,
		&k.RateLimit,
		&k.IsActive,
		&k.LastUsedAt,
		&k.ExpiresAt,
		&k.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("api key not found")
		}
		return nil, fmt.Errorf("scan api key: %w", err)
	}
	return &k, nil
}
