-- +goose Up
ALTER TABLE kyc_verifications ADD COLUMN phone_hash VARCHAR(64) NOT NULL DEFAULT '';
UPDATE kyc_verifications SET phone_hash = encode(sha256(phone::bytea), 'hex') WHERE phone_hash = '';
CREATE INDEX IF NOT EXISTS idx_kyc_phone_hash ON kyc_verifications(phone_hash);
DROP INDEX IF EXISTS idx_kyc_phone;
CREATE INDEX IF NOT EXISTS idx_kyc_phone_hash_country ON kyc_verifications(phone_hash, country_code);

-- +goose Down
DROP INDEX IF EXISTS idx_kyc_phone_hash_country;
DROP INDEX IF EXISTS idx_kyc_phone_hash;
ALTER TABLE kyc_verifications DROP COLUMN phone_hash;
CREATE INDEX IF NOT EXISTS idx_kyc_phone ON kyc_verifications(phone);
