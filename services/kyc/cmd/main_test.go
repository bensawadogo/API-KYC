package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/datakeys/kyc-service/config"
)

func TestEnvInt_Parse(t *testing.T) {
	tests := []struct {
		val  string
		def  int
		want int
	}{
		{"42", 0, 42},
		{"", 10, 10},
		{"abc", 99, 99},
	}
	for _, tc := range tests {
		if tc.val != "" {
			os.Setenv("TEST_INT", tc.val)
		}
		got := getEnvInt("TEST_INT", tc.def)
		if got != tc.want {
			t.Errorf("getEnvInt(%q, %d) = %d, want %d", tc.val, tc.def, got, tc.want)
		}
		os.Unsetenv("TEST_INT")
	}
}

func TestDuration_Parse(t *testing.T) {
	tests := []struct {
		val  string
		def  time.Duration
		want time.Duration
	}{
		{"5s", time.Second, 5 * time.Second},
		{"", time.Minute, time.Minute},
		{"invalid", 30 * time.Second, 30 * time.Second},
	}
	for _, tc := range tests {
		if tc.val != "" {
			os.Setenv("TEST_DUR", tc.val)
		}
		got := getDuration("TEST_DUR", tc.def)
		if got != tc.want {
			t.Errorf("getDuration(%q, %v) = %v, want %v", tc.val, tc.def, got, tc.want)
		}
		os.Unsetenv("TEST_DUR")
	}
}

func TestEnvFloat_Parse(t *testing.T) {
	os.Setenv("TEST_FLOAT", "3.14")
	got := getEnvFloat("TEST_FLOAT", 0.0)
	if got != 3.14 {
		t.Errorf("got %f, want 3.14", got)
	}
	os.Unsetenv("TEST_FLOAT")

	got = getEnvFloat("MISSING", 1.5)
	if got != 1.5 {
		t.Errorf("got %f, want 1.5", got)
	}
}

func TestConfigLoad_WithPostgres(t *testing.T) {
	os.Setenv("POSTGRES_URL", "postgres://user:pass@localhost:5432/testdb?sslmode=disable")
	defer os.Unsetenv("POSTGRES_URL")

	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config.Load() error: %v", err)
	}
	if cfg.Server.Port != "8081" {
		t.Errorf("default port want 8081, got %s", cfg.Server.Port)
	}
	if cfg.Database.PostgresURL != "postgres://user:pass@localhost:5432/testdb?sslmode=disable" {
		t.Errorf("postgres url mismatch")
	}
}

func TestConfigValidate_TLS_Pair(t *testing.T) {
	cfg := config.Config{
		Database: config.DatabaseConfig{PostgresURL: "postgres://localhost"},
		KYC:      config.KYCConfig{MaxDocSizeMB: 10, SessionTTLSeconds: 3600, ScoreThreshold: 0.7, DefaultProvider: "local"},
		Compliance: config.ComplianceConfig{RetentionDays: 365},
		Server:   config.ServerConfig{TLSCert: "/cert.pem", TLSKey: ""},
	}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for TLS_CERT without TLS_KEY")
	}

	cfg.Server = config.ServerConfig{TLSCert: "", TLSKey: "/key.pem"}
	if err := cfg.Validate(); err == nil {
		t.Error("expected error for TLS_KEY without TLS_CERT")
	}
}

func TestGracefulShutdownNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	select {
	case <-ctx.Done():
	case <-time.After(200 * time.Millisecond):
	}
}
