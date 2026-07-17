package audit

import (
	"context"
	"runtime"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestLogAsync_NoBlock(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	al := &AuditLogger{pool: nil, logger: logger}

	before := runtime.NumGoroutine()
	entry := AuditEntry{
		EventType:      EventInitiated,
		VerificationID: "test-id",
		Phone:          "+22670000000",
		CountryCode:    "BF",
	}
	al.LogAsync(context.Background(), entry)
	time.Sleep(100 * time.Millisecond)
	after := runtime.NumGoroutine()

	if after > before+2 {
		t.Errorf("possible goroutine leak: before=%d after=%d", before, after)
	}
}

func TestHashPhone_Deterministic(t *testing.T) {
	phone := "+22670000000"
	h1 := hashPhone(phone)
	h2 := hashPhone(phone)
	if h1 != h2 {
		t.Error("hash should be deterministic")
	}
	if len(h1) != 64 {
		t.Errorf("expected 64 hex chars, got %d", len(h1))
	}
}

func TestEventTypes_Constants(t *testing.T) {
	events := []string{
		EventInitiated, EventProcessing, EventApproved, EventRejected,
		EventManualReview, EventExpired, EventWebhookSent, EventWebhookFail,
		EventProviderFail, EventDocDeleted,
	}

	seen := make(map[string]bool)
	for _, e := range events {
		if e == "" {
			t.Error("event type should not be empty")
		}
		if seen[e] {
			t.Errorf("duplicate event type: %s", e)
		}
		seen[e] = true
	}
}

func TestAuditEntry_Fields(t *testing.T) {
	entry := AuditEntry{
		EventType:      EventInitiated,
		VerificationID: "abc-123",
		Phone:          "+22670000000",
		CountryCode:    "BF",
		DocType:        "NATIONAL_ID",
		Provider:       "smileid",
		StatusAfter:    "pending",
		Score:          0.95,
		Flags:          []string{},
		IPAddress:      "127.0.0.1",
		UserAgent:      "test",
		DurationMS:     150,
		Metadata:       map[string]interface{}{"key": "val"},
	}

	if entry.EventType != EventInitiated {
		t.Error("event type mismatch")
	}
	if entry.Phone != "+22670000000" {
		t.Error("phone mismatch")
	}
}

func TestLog_FailOpen(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	al := &AuditLogger{pool: nil, logger: logger}

	entry := AuditEntry{
		EventType:      EventInitiated,
		VerificationID: "test",
		Phone:          "+22670000000",
		CountryCode:    "BF",
	}

	// Should not panic with nil pool
	al.Log(context.Background(), entry)
}