package model

import "time"

const (
	ScopeKYCInitiate = "kyc:initiate"
	ScopeKYCStatus   = "kyc:status"
	ScopeKYCAdmin    = "kyc:admin"
	ScopeKYCWebhook  = "kyc:webhook"
)

type APIKey struct {
	ID         string     `db:"id"`
	ClientName string     `db:"client_name"`
	KeyHash    string     `db:"key_hash"`
	KeyPrefix  string     `db:"key_prefix"`
	Scopes     []string   `db:"scopes"`
	RateLimit  int        `db:"rate_limit"`
	IsActive   bool       `db:"is_active"`
	LastUsedAt *time.Time `db:"last_used_at"`
	ExpiresAt  *time.Time `db:"expires_at"`
	CreatedAt  time.Time  `db:"created_at"`
}
