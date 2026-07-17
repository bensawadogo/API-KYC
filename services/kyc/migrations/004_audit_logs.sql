-- +goose Up
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

CREATE INDEX IF NOT EXISTS idx_audit_verification_id
  ON kyc_audit_log(verification_id);
CREATE INDEX IF NOT EXISTS idx_audit_phone_hash
  ON kyc_audit_log(phone_hash);
CREATE INDEX IF NOT EXISTS idx_audit_created_at
  ON kyc_audit_log(created_at);

-- +goose Down
DROP TABLE IF EXISTS kyc_audit_log;
