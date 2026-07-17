package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Server        ServerConfig
	Database      DatabaseConfig
	Providers     ProvidersConfig
	KYC           KYCConfig
	Compliance    ComplianceConfig
	AML           AMLConfig
	Storage       StorageConfig
	Webhooks      WebhookConfig
	Observability ObservabilityConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	PostgresURL string
}

type ProvidersConfig struct {
	SmileID   SmileIDConfig
	Youverify YouverifyConfig
	SumSub    SumSubConfig
}

type SmileIDConfig struct {
	ApiKey      string
	PartnerId   string
	CallbackUrl string
	Sandbox     bool
}

type YouverifyConfig struct {
	ApiKey  string
	BaseURL string
	Sandbox bool
}

type SumSubConfig struct {
	AppToken  string
	SecretKey string
	BaseURL   string
	Sandbox   bool
}

type KYCConfig struct {
	MaxDocSizeMB      int
	SessionTTLSeconds int
	ScoreThreshold    float64
	DefaultProvider   string
}

type ComplianceConfig struct {
	ConsentRequired bool
	RetentionDays   int
	AuditEnabled    bool
}

type AMLConfig struct {
	Provider    string
	APIKey      string
	BaseURL     string
	Threshold   float64
	BlockOnFail bool
}

type StorageConfig struct {
	Provider  string
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

type WebhookConfig struct {
	SmileIDAllowedCIDRs []string
	SumSubAllowedCIDRs  []string
}

type ObservabilityConfig struct {
	TracingEnabled bool
	OTLPEndpoint   string
}

func Load() (Config, error) {
	cfg := Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8081"),
			Env:  getEnv("SERVER_ENV", "development"),
		},
		Database: DatabaseConfig{
			PostgresURL: os.Getenv("POSTGRES_URL"),
		},
		Providers: ProvidersConfig{
			SmileID: SmileIDConfig{
				ApiKey:      os.Getenv("SMILEID_API_KEY"),
				PartnerId:   os.Getenv("SMILEID_PARTNER_ID"),
				CallbackUrl: os.Getenv("SMILEID_CALLBACK_URL"),
				Sandbox:     getEnvBool("SMILEID_SANDBOX", true),
			},
			Youverify: YouverifyConfig{
				ApiKey:  os.Getenv("YOUVERIFY_API_KEY"),
				BaseURL: getEnv("YOUVERIFY_BASE_URL", ""),
				Sandbox: getEnvBool("YOUVERIFY_SANDBOX", true),
			},
			SumSub: SumSubConfig{
				AppToken:  os.Getenv("SUMSUB_APP_TOKEN"),
				SecretKey: os.Getenv("SUMSUB_SECRET_KEY"),
				BaseURL:   getEnv("SUMSUB_BASE_URL", "https://api.sumsub.com"),
				Sandbox:   getEnvBool("SUMSUB_SANDBOX", true),
			},
		},
		KYC: KYCConfig{
			MaxDocSizeMB:      getEnvInt("KYC_MAX_DOC_SIZE_MB", 10),
			SessionTTLSeconds: getEnvInt("KYC_SESSION_TTL_SECONDS", 3600),
			ScoreThreshold:    getEnvFloat("KYC_SCORE_THRESHOLD", 0.70),
			DefaultProvider:   getEnv("KYC_DEFAULT_PROVIDER", "smileid"),
		},
		Compliance: ComplianceConfig{
			ConsentRequired: getEnvBool("COMPLIANCE_CONSENT_REQUIRED", true),
			RetentionDays:   getEnvInt("COMPLIANCE_RETENTION_DAYS", 1825),
			AuditEnabled:    getEnvBool("COMPLIANCE_AUDIT_ENABLED", true),
		},
		AML: AMLConfig{
			Provider:    getEnv("AML_PROVIDER", "local"),
			APIKey:      os.Getenv("OPENSANCTIONS_API_KEY"),
			BaseURL:     getEnv("AML_BASE_URL", "https://api.opensanctions.org"),
			Threshold:   getEnvFloat("AML_THRESHOLD", 0.70),
			BlockOnFail: getEnvBool("AML_BLOCK_ON_FAIL", false),
		},
		Storage: StorageConfig{
			Provider:  getEnv("STORAGE_PROVIDER", "memory"),
			Endpoint:  getEnv("MINIO_ENDPOINT", "localhost:9000"),
			AccessKey: os.Getenv("MINIO_ACCESS_KEY"),
			SecretKey: os.Getenv("MINIO_SECRET_KEY"),
			Bucket:    getEnv("MINIO_BUCKET", "kyc-documents"),
			UseSSL:    getEnvBool("MINIO_USE_SSL", false),
		},
		Webhooks: WebhookConfig{
			SmileIDAllowedCIDRs: getEnvCSV("SMILEID_WEBHOOK_CIDRS"),
			SumSubAllowedCIDRs:  getEnvCSV("SUMSUB_WEBHOOK_CIDRS"),
		},
		Observability: ObservabilityConfig{
			TracingEnabled: getEnvBool("TRACING_ENABLED", false),
			OTLPEndpoint:   getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		},
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	if strings.TrimSpace(c.Database.PostgresURL) == "" {
		return fmt.Errorf("POSTGRES_URL is required")
	}

	if c.KYC.MaxDocSizeMB <= 0 {
		return fmt.Errorf("KYC_MAX_DOC_SIZE_MB must be positive")
	}

	if c.KYC.SessionTTLSeconds <= 0 {
		return fmt.Errorf("KYC_SESSION_TTL_SECONDS must be positive")
	}

	if c.KYC.ScoreThreshold < 0 || c.KYC.ScoreThreshold > 1 {
		return fmt.Errorf("KYC_SCORE_THRESHOLD must be between 0 and 1")
	}

	validProviders := map[string]bool{
		"smileid": true, "youverify": true, "sumsub": true, "local": true,
	}
	if !validProviders[c.KYC.DefaultProvider] {
		return fmt.Errorf("KYC_DEFAULT_PROVIDER must be one of: smileid, youverify, sumsub, local")
	}

	if c.Compliance.RetentionDays <= 0 {
		return fmt.Errorf("COMPLIANCE_RETENTION_DAYS must be positive")
	}

	return nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getEnvFloat(key string, fallback float64) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return fallback
	}
	return f
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func getEnvCSV(key string) []string {
	val := os.Getenv(key)
	if val == "" {
		return []string{}
	}
	parts := strings.Split(val, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			result = append(result, t)
		}
	}
	return result
}
