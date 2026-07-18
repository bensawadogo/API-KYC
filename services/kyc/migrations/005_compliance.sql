-- +goose Up
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

CREATE TABLE IF NOT EXISTS kyc_sar_reports (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  verification_id   UUID NOT NULL REFERENCES kyc_verifications(id),
  entity_name       VARCHAR(200) NOT NULL,
  country_code      CHAR(2) NOT NULL,
  doc_type          VARCHAR(30),
  risk_score        DECIMAL(4,3),
  flags             TEXT[],
  aml_match_details TEXT,
  regulation        VARCHAR(50),
  narrative         TEXT,
  filing_status     VARCHAR(20) NOT NULL DEFAULT 'draft',
  filed_at          TIMESTAMPTZ,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_sar_verification ON kyc_sar_reports(verification_id);
CREATE INDEX IF NOT EXISTS idx_sar_status ON kyc_sar_reports(filing_status);
CREATE INDEX IF NOT EXISTS idx_sar_created ON kyc_sar_reports(created_at);

-- +goose Down
DROP TABLE IF EXISTS kyc_sar_reports CASCADE;
DROP TABLE IF EXISTS kyc_consent_records CASCADE;
