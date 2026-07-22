package datakeys

import (
	"context"
	"testing"
)

func TestNew_EmptyAPIKey(t *testing.T) {
	_, err := New("")
	if err == nil {
		t.Fatal("expected error for empty api key")
	}
	if err.Error() != "[KYC_AUTH_001] API key manquante (HTTP 401)" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNew_SandboxKey(t *testing.T) {
	dk, err := New("dk_test_fakekey")
	if err != nil {
		t.Fatal(err)
	}
	if dk.Livemode {
		t.Error("expected livemode=false for dk_test_ key")
	}
}

func TestNew_LiveKey(t *testing.T) {
	dk, err := New("dk_live_fakekey")
	if err != nil {
		t.Fatal(err)
	}
	if !dk.Livemode {
		t.Error("expected livemode=true for dk_live_ key")
	}
}

func TestMustNew_PanicsOnEmpty(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic")
		}
	}()
	MustNew("")
}

func TestKYCService_Retrieve_EmptyID(t *testing.T) {
	dk, _ := New("dk_test_fakekey")
	_, err := dk.KYC.Retrieve(context.Background(), "")
	if err == nil || err.Error() != "verificationID requis" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestKYCService_WaitForCompletion_Timeout(t *testing.T) {
	// Juste vérifier que la fonction ne panique pas
	// et qu'elle retourne une erreur de timeout
	dk, _ := New("dk_test_fakekey")
	// Le timeout est détecté car l'appel API échoue (pas de serveur)
	ctx := context.Background()
	_, err := dk.KYC.WaitForCompletion(ctx, "ver_fake", 0)
	if err == nil {
		t.Error("expected error")
	}
}

func TestExtractCode(t *testing.T) {
	tests := []struct {
		input string
		want  ErrorCode
	}{
		{"KYC_AUTH_001: bad key", ErrAuthMissing},
		{"KYC_RATE_001: slow down", ErrRateLimit},
		{"random error", ErrUnknown},
		{"", ErrUnknown},
	}
	for _, tt := range tests {
		got := extractCode(tt.input)
		if got != tt.want {
			t.Errorf("extractCode(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
