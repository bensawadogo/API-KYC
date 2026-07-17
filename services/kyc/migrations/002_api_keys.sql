-- +goose Up
CREATE TABLE IF NOT EXISTS api_keys (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  client_name  VARCHAR(100) NOT NULL,
  key_hash     VARCHAR(255) NOT NULL UNIQUE,
  key_prefix   VARCHAR(20) NOT NULL,
  scopes       TEXT[] NOT NULL DEFAULT '{}',
  rate_limit   INT NOT NULL DEFAULT 60,
  is_active    BOOLEAN NOT NULL DEFAULT true,
  last_used_at TIMESTAMPTZ,
  expires_at   TIMESTAMPTZ,
  created_at   TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_apikeys_prefix
  ON api_keys(key_prefix);
CREATE INDEX IF NOT EXISTS idx_apikeys_active
  ON api_keys(is_active) WHERE is_active = true;

-- +goose Down
DROP TABLE IF EXISTS api_keys;
