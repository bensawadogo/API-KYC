package config

import (
	"os"
	"testing"
)

func TestGetEnv_WithValue(t *testing.T) {
	os.Setenv("TEST_KEY", "custom")
	defer os.Unsetenv("TEST_KEY")
	if v := getEnv("TEST_KEY", "fallback"); v != "custom" {
		t.Errorf("expected custom, got %s", v)
	}
}

func TestGetEnv_Fallback(t *testing.T) {
	os.Unsetenv("TEST_MISSING")
	if v := getEnv("TEST_MISSING", "default"); v != "default" {
		t.Errorf("expected default, got %s", v)
	}
}

func TestGetEnvInt_WithValue(t *testing.T) {
	os.Setenv("TEST_INT", "42")
	defer os.Unsetenv("TEST_INT")
	if v := getEnvInt("TEST_INT", 1); v != 42 {
		t.Errorf("expected 42, got %d", v)
	}
}

func TestGetEnvInt_Fallback(t *testing.T) {
	os.Unsetenv("TEST_INT_MISSING")
	if v := getEnvInt("TEST_INT_MISSING", 99); v != 99 {
		t.Errorf("expected 99, got %d", v)
	}
}

func TestGetEnvInt_Invalid(t *testing.T) {
	os.Setenv("TEST_INT_INV", "notanumber")
	defer os.Unsetenv("TEST_INT_INV")
	if v := getEnvInt("TEST_INT_INV", 77); v != 77 {
		t.Errorf("expected fallback 77, got %d", v)
	}
}

func TestGetEnvFloat_WithValue(t *testing.T) {
	os.Setenv("TEST_FLOAT", "0.85")
	defer os.Unsetenv("TEST_FLOAT")
	if v := getEnvFloat("TEST_FLOAT", 0.5); v != 0.85 {
		t.Errorf("expected 0.85, got %f", v)
	}
}

func TestGetEnvFloat_Fallback(t *testing.T) {
	os.Unsetenv("TEST_FLOAT_MISSING")
	if v := getEnvFloat("TEST_FLOAT_MISSING", 0.75); v != 0.75 {
		t.Errorf("expected 0.75, got %f", v)
	}
}

func TestGetEnvFloat_Invalid(t *testing.T) {
	os.Setenv("TEST_FLOAT_INV", "bad")
	defer os.Unsetenv("TEST_FLOAT_INV")
	if v := getEnvFloat("TEST_FLOAT_INV", 0.6); v != 0.6 {
		t.Errorf("expected fallback 0.6, got %f", v)
	}
}

func TestGetEnvBool_WithTrue(t *testing.T) {
	os.Setenv("TEST_BOOL", "true")
	defer os.Unsetenv("TEST_BOOL")
	if v := getEnvBool("TEST_BOOL", false); v != true {
		t.Error("expected true")
	}
}

func TestGetEnvBool_WithFalse(t *testing.T) {
	os.Setenv("TEST_BOOL_F", "false")
	defer os.Unsetenv("TEST_BOOL_F")
	if v := getEnvBool("TEST_BOOL_F", true); v != false {
		t.Error("expected false")
	}
}

func TestGetEnvBool_Fallback(t *testing.T) {
	os.Unsetenv("TEST_BOOL_MISSING")
	if v := getEnvBool("TEST_BOOL_MISSING", true); v != true {
		t.Error("expected true fallback")
	}
}

func TestGetEnvBool_Invalid(t *testing.T) {
	os.Setenv("TEST_BOOL_INV", "maybe")
	defer os.Unsetenv("TEST_BOOL_INV")
	if v := getEnvBool("TEST_BOOL_INV", false); v != false {
		t.Error("expected fallback false")
	}
}

func TestValidate_EmptyPostgresURL(t *testing.T) {
	cfg := Config{
		KYC: KYCConfig{MaxDocSizeMB: 10, SessionTTLSeconds: 3600, ScoreThreshold: 0.7, DefaultProvider: "smileid"},
		Compliance: ComplianceConfig{RetentionDays: 365},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for empty POSTGRES_URL")
	}
}

func TestValidate_InvalidDocSize(t *testing.T) {
	cfg := Config{
		Database: DatabaseConfig{PostgresURL: "postgres://localhost"},
		KYC:      KYCConfig{MaxDocSizeMB: 0, SessionTTLSeconds: 3600, ScoreThreshold: 0.7, DefaultProvider: "smileid"},
		Compliance: ComplianceConfig{RetentionDays: 365},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for zero doc size")
	}
}

func TestValidate_InvalidSessionTTL(t *testing.T) {
	cfg := Config{
		Database: DatabaseConfig{PostgresURL: "postgres://localhost"},
		KYC:      KYCConfig{MaxDocSizeMB: 10, SessionTTLSeconds: 0, ScoreThreshold: 0.7, DefaultProvider: "smileid"},
		Compliance: ComplianceConfig{RetentionDays: 365},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for zero session TTL")
	}
}

func TestValidate_InvalidScoreThreshold(t *testing.T) {
	cfg := Config{
		Database: DatabaseConfig{PostgresURL: "postgres://localhost"},
		KYC:      KYCConfig{MaxDocSizeMB: 10, SessionTTLSeconds: 3600, ScoreThreshold: 1.5, DefaultProvider: "smileid"},
		Compliance: ComplianceConfig{RetentionDays: 365},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for score > 1")
	}
}

func TestValidate_InvalidProvider(t *testing.T) {
	cfg := Config{
		Database: DatabaseConfig{PostgresURL: "postgres://localhost"},
		KYC:      KYCConfig{MaxDocSizeMB: 10, SessionTTLSeconds: 3600, ScoreThreshold: 0.7, DefaultProvider: "unknown"},
		Compliance: ComplianceConfig{RetentionDays: 365},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for invalid provider")
	}
}

func TestValidate_InvalidRetentionDays(t *testing.T) {
	cfg := Config{
		Database: DatabaseConfig{PostgresURL: "postgres://localhost"},
		KYC:      KYCConfig{MaxDocSizeMB: 10, SessionTTLSeconds: 3600, ScoreThreshold: 0.7, DefaultProvider: "smileid"},
		Compliance: ComplianceConfig{RetentionDays: 0},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for zero retention days")
	}
}

func TestValidate_Valid(t *testing.T) {
	cfg := Config{
		Database: DatabaseConfig{PostgresURL: "postgres://localhost"},
		KYC:      KYCConfig{MaxDocSizeMB: 10, SessionTTLSeconds: 3600, ScoreThreshold: 0.7, DefaultProvider: "local"},
		Compliance: ComplianceConfig{RetentionDays: 365},
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestLoad_Defaults(t *testing.T) {
	os.Setenv("POSTGRES_URL", "postgres://test")
	defer os.Unsetenv("POSTGRES_URL")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Port != "8081" {
		t.Errorf("expected 8081, got %s", cfg.Server.Port)
	}
	if cfg.KYC.MaxDocSizeMB != 10 {
		t.Errorf("expected 10, got %d", cfg.KYC.MaxDocSizeMB)
	}
	if cfg.KYC.ScoreThreshold != 0.70 {
		t.Errorf("expected 0.70, got %f", cfg.KYC.ScoreThreshold)
	}
	if cfg.Providers.SmileID.Sandbox != true {
		t.Error("expected sandbox true by default")
	}
	if cfg.Compliance.ConsentRequired != true {
		t.Error("expected consent required by default")
	}
	if cfg.Compliance.AuditEnabled != true {
		t.Error("expected audit enabled by default")
	}
}

func TestLoad_FromEnv(t *testing.T) {
	os.Setenv("POSTGRES_URL", "postgres://prod")
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("SERVER_ENV", "production")
	os.Setenv("KYC_MAX_DOC_SIZE_MB", "20")
	os.Setenv("KYC_SCORE_THRESHOLD", "0.85")
	os.Setenv("SMILEID_SANDBOX", "false")
	os.Setenv("SUMSUB_SANDBOX", "false")
	os.Setenv("YOUVERIFY_SANDBOX", "false")
	os.Setenv("COMPLIANCE_CONSENT_REQUIRED", "false")
	os.Setenv("COMPLIANCE_AUDIT_ENABLED", "false")
	os.Setenv("KYC_DEFAULT_PROVIDER", "youverify")
	os.Setenv("KYC_SESSION_TTL_SECONDS", "7200")
	os.Setenv("COMPLIANCE_RETENTION_DAYS", "90")
	defer func() {
		os.Unsetenv("POSTGRES_URL")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("SERVER_ENV")
		os.Unsetenv("KYC_MAX_DOC_SIZE_MB")
		os.Unsetenv("KYC_SCORE_THRESHOLD")
		os.Unsetenv("SMILEID_SANDBOX")
		os.Unsetenv("SUMSUB_SANDBOX")
		os.Unsetenv("YOUVERIFY_SANDBOX")
		os.Unsetenv("COMPLIANCE_CONSENT_REQUIRED")
		os.Unsetenv("COMPLIANCE_AUDIT_ENABLED")
		os.Unsetenv("KYC_DEFAULT_PROVIDER")
		os.Unsetenv("KYC_SESSION_TTL_SECONDS")
		os.Unsetenv("COMPLIANCE_RETENTION_DAYS")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Server.Port != "9090" {
		t.Errorf("expected 9090, got %s", cfg.Server.Port)
	}
	if cfg.Server.Env != "production" {
		t.Errorf("expected production, got %s", cfg.Server.Env)
	}
	if cfg.KYC.MaxDocSizeMB != 20 {
		t.Errorf("expected 20, got %d", cfg.KYC.MaxDocSizeMB)
	}
	if cfg.KYC.ScoreThreshold != 0.85 {
		t.Errorf("expected 0.85, got %f", cfg.KYC.ScoreThreshold)
	}
	if cfg.Providers.SmileID.Sandbox != false {
		t.Error("expected sandbox false")
	}
	if cfg.Providers.Youverify.Sandbox != false {
		t.Error("expected youverify sandbox false")
	}
	if cfg.Providers.SumSub.Sandbox != false {
		t.Error("expected sumsub sandbox false")
	}
	if cfg.Compliance.ConsentRequired != false {
		t.Error("expected consent required false")
	}
	if cfg.Compliance.AuditEnabled != false {
		t.Error("expected audit enabled false")
	}
	if cfg.KYC.DefaultProvider != "youverify" {
		t.Errorf("expected youverify, got %s", cfg.KYC.DefaultProvider)
	}
	if cfg.KYC.SessionTTLSeconds != 7200 {
		t.Errorf("expected 7200, got %d", cfg.KYC.SessionTTLSeconds)
	}
	if cfg.Compliance.RetentionDays != 90 {
		t.Errorf("expected 90, got %d", cfg.Compliance.RetentionDays)
	}
}

func TestLoad_MissingPostgres(t *testing.T) {
	os.Unsetenv("POSTGRES_URL")
	_, err := Load()
	if err == nil {
		t.Error("expected error for missing POSTGRES_URL")
	}
}
