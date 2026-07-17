-- +goose Up
CREATE TABLE IF NOT EXISTS kyc_aml_results (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  verification_id  UUID NOT NULL REFERENCES kyc_verifications(id),
  is_sanctioned    BOOLEAN NOT NULL DEFAULT false,
  is_pep           BOOLEAN NOT NULL DEFAULT false,
  aml_score        DECIMAL(4,3),
  matches_count    INT DEFAULT 0,
  source           VARCHAR(50),
  screened_at      TIMESTAMPTZ NOT NULL,
  raw_matches      JSONB,
  created_at       TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_aml_verification
  ON kyc_aml_results(verification_id);
CREATE INDEX IF NOT EXISTS idx_aml_sanctioned
  ON kyc_aml_results(is_sanctioned) WHERE is_sanctioned = true;

-- +goose Down
DROP TABLE IF EXISTS kyc_aml_results;
