-- +goose Up
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS kyc_verifications (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  phone           VARCHAR(20) NOT NULL,
  country_code    CHAR(2) NOT NULL,
  doc_type        VARCHAR(30) NOT NULL,
  status          VARCHAR(20) NOT NULL DEFAULT 'pending',
  score           DECIMAL(4,3),
  provider        VARCHAR(50),
  flags           TEXT[],
  callback_url    TEXT,
  consent         BOOLEAN NOT NULL DEFAULT false,
  consent_at      TIMESTAMPTZ DEFAULT NOW(),
  processed_at    TIMESTAMPTZ,
  expires_at      TIMESTAMPTZ NOT NULL,
  created_at      TIMESTAMPTZ DEFAULT NOW(),
  updated_at      TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_kyc_phone
  ON kyc_verifications(phone);
CREATE INDEX IF NOT EXISTS idx_kyc_status
  ON kyc_verifications(status);
CREATE INDEX IF NOT EXISTS idx_kyc_country
  ON kyc_verifications(country_code);
CREATE INDEX IF NOT EXISTS idx_kyc_phone_country
  ON kyc_verifications(phone, country_code);
CREATE INDEX IF NOT EXISTS idx_kyc_expires
  ON kyc_verifications(expires_at)
  WHERE status IN ('pending','processing');

-- +goose Down
DROP TABLE IF EXISTS kyc_verifications CASCADE;
