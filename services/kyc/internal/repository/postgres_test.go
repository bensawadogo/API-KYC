package repository_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/datakeys/kyc-service/internal"
	"github.com/datakeys/kyc-service/internal/model"
	"github.com/datakeys/kyc-service/internal/repository"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestDBURL(t *testing.T) string {
	t.Helper()
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("TEST_DATABASE_URL not set: skipping integration test")
	}
	return url
}

func setupTestRepo(t *testing.T) (*repository.PostgresRepository, context.Context) {
	t.Helper()
	ctx := context.Background()
	url := getTestDBURL(t)
	repo, err := repository.NewPostgresRepository(ctx, url)
	require.NoError(t, err)
	t.Cleanup(repo.Close)

	err = repository.RunMigrations(repo.Pool())
	require.NoError(t, err)

	return repo, ctx
}

func TestRepository_CreateAndGetVerification(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	v := &model.VerificationResult{
		VerificationID: uuid.New().String(),
		Phone:          "+221771234567",
		CountryCode:    "SN",
		DocType:        "NATIONAL_ID",
		Status:         model.StatusPending,
		Score:          0.0,
		Provider:       "smileid",
		Consent:        true,
		ExpiresAt:      time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
	}

	err := repo.CreateVerification(ctx, v)
	require.NoError(t, err)

	got, err := repo.GetVerification(ctx, v.VerificationID)
	require.NoError(t, err)
	assert.Equal(t, v.VerificationID, got.VerificationID)
	assert.Equal(t, v.Phone, got.Phone)
	assert.Equal(t, v.CountryCode, got.CountryCode)
	assert.Equal(t, v.Status, got.Status)
	assert.True(t, got.Consent)
}

func TestRepository_CreateAndGetVerification_NotFound(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	_, err := repo.GetVerification(ctx, uuid.New().String())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRepository_UpdateStatus(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	v := &model.VerificationResult{
		VerificationID: uuid.New().String(),
		Phone:          "+2250102030405",
		CountryCode:    "CI",
		DocType:        "PASSPORT",
		Status:         model.StatusPending,
		Consent:        true,
		ExpiresAt:      time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
	}

	err := repo.CreateVerification(ctx, v)
	require.NoError(t, err)

	err = repo.UpdateStatus(ctx, v.VerificationID, model.StatusApproved, 0.95,
		[]string{}, "smileid")
	require.NoError(t, err)

	got, err := repo.GetVerification(ctx, v.VerificationID)
	require.NoError(t, err)
	assert.Equal(t, model.StatusApproved, got.Status)
	assert.Equal(t, 0.95, got.Score)
	assert.Equal(t, "smileid", got.Provider)
}

func TestRepository_ListByPhone(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	phone := "+233501234567"
	for i := 0; i < 3; i++ {
		v := &model.VerificationResult{
			VerificationID: uuid.New().String(),
			Phone:          phone,
			CountryCode:    "GH",
			DocType:        "NATIONAL_ID",
			Status:         model.StatusApproved,
			Consent:        true,
			ExpiresAt:      time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
		}
		err := repo.CreateVerification(ctx, v)
		require.NoError(t, err)
	}

	results, err := repo.ListByPhone(ctx, phone, 10)
	require.NoError(t, err)
	assert.Len(t, results, 3)
	for _, r := range results {
		assert.Equal(t, phone, r.Phone)
	}
}

func TestRepository_ListByPhone_DefaultLimit(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	phone := "+226701234567"
	for i := 0; i < 5; i++ {
		v := &model.VerificationResult{
			VerificationID: uuid.New().String(),
			Phone:          phone,
			CountryCode:    "BF",
			DocType:        "NATIONAL_ID",
			Status:         model.StatusPending,
			Consent:        true,
			ExpiresAt:      time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
		}
		err := repo.CreateVerification(ctx, v)
		require.NoError(t, err)
	}

	results, err := repo.ListByPhone(ctx, phone, 0)
	require.NoError(t, err)
	assert.Len(t, results, 5)
}

func TestRepository_ListByPhone_NoResults(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	results, err := repo.ListByPhone(ctx, "+999000000000", 10)
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestRepository_ExistsApproved_True(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	phone := "+224621123456"
	v := &model.VerificationResult{
		VerificationID: uuid.New().String(),
		Phone:          phone,
		CountryCode:    "GN",
		DocType:        "NATIONAL_ID",
		Status:         model.StatusApproved,
		Score:          0.92,
		Provider:       "smileid",
		Consent:        true,
		ExpiresAt:      time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
	}
	err := repo.CreateVerification(ctx, v)
	require.NoError(t, err)

	exists, err := repo.ExistsApproved(ctx, phone, "GN")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestRepository_ExistsApproved_FalseWhenExpired(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	phone := "+223701234567"
	v := &model.VerificationResult{
		VerificationID: uuid.New().String(),
		Phone:          phone,
		CountryCode:    "ML",
		DocType:        "NATIONAL_ID",
		Status:         model.StatusApproved,
		Provider:       "smileid",
		Consent:        true,
		ExpiresAt:      time.Now().UTC().Add(-1 * time.Hour).Format(time.RFC3339),
	}
	err := repo.CreateVerification(ctx, v)
	require.NoError(t, err)

	exists, err := repo.ExistsApproved(ctx, phone, "ML")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestRepository_ExistsApproved_DifferentCountry(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	phone := "+221771234567"
	v := &model.VerificationResult{
		VerificationID: uuid.New().String(),
		Phone:          phone,
		CountryCode:    "SN",
		DocType:        "NATIONAL_ID",
		Status:         model.StatusApproved,
		Provider:       "smileid",
		Consent:        true,
		ExpiresAt:      time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339),
	}
	err := repo.CreateVerification(ctx, v)
	require.NoError(t, err)

	exists, err := repo.ExistsApproved(ctx, phone, "ML")
	require.NoError(t, err)
	assert.False(t, exists)
}

func TestRepository_SaveAndGetAMLResult(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	verificationID := uuid.New().String()
	v := &model.VerificationResult{
		VerificationID: verificationID,
		Phone:          "+2250102030405",
		CountryCode:    "CI",
		DocType:        "PASSPORT",
		Status:         model.StatusPending,
		Consent:        true,
		ExpiresAt:      time.Now().UTC().Add(time.Hour).Format(time.RFC3339),
	}
	err := repo.CreateVerification(ctx, v)
	require.NoError(t, err)

	amlResult := &internal.AMLResult{
		IsSanctioned: true,
		IsPEP:        false,
		Score:        0.95,
		Matches: []internal.AMLMatch{
			{EntityName: "John Doe", EntityID: "SDN-12345", Topics: []string{"sanction"}, Score: 0.98, Dataset: "OFAC"},
		},
		ScreenedAt: time.Now().UTC(),
		Source:     "opensanctions",
	}

	err = repo.SaveAMLResult(ctx, verificationID, amlResult)
	require.NoError(t, err)

	saved, err := repo.GetAMLResult(ctx, verificationID)
	require.NoError(t, err)
	assert.True(t, saved.IsSanctioned)
	assert.False(t, saved.IsPEP)
	assert.Equal(t, 0.95, saved.Score)
	assert.Len(t, saved.Matches, 1)
	assert.Equal(t, "John Doe", saved.Matches[0].EntityName)
	assert.Equal(t, "opensanctions", saved.Source)
}

func TestRepository_SaveAMLResult_NotFound(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	_, err := repo.GetAMLResult(ctx, uuid.New().String())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRepository_FindAPIKeyByPrefix(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	key := &model.APIKey{
		ClientName: "test-client",
		KeyHash:    "abc123def456",
		KeyPrefix:  "dk_test_pref",
		Scopes:     []string{"kyc:initiate"},
		RateLimit:  60,
		IsActive:   true,
	}
	err := repo.CreateAPIKey(ctx, key)
	require.NoError(t, err)

	found, err := repo.FindByPrefix(ctx, "dk_test_pref")
	require.NoError(t, err)
	assert.Equal(t, "test-client", found.ClientName)
	assert.Equal(t, "dk_test_pref", found.KeyPrefix)
	assert.Equal(t, []string{"kyc:initiate"}, found.Scopes)
}

func TestRepository_FindAPIKeyByPrefix_NotFound(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	_, err := repo.FindByPrefix(ctx, "dk_nonexist")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRepository_ValidateKey_Success(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	rawKey := "dk_test_key_integration_test_123"
	prefix := rawKey[:12]

	sum := sha256.Sum256([]byte(rawKey))

	key := &model.APIKey{
		ClientName: "validate-test",
		KeyHash:    hex.EncodeToString(sum[:]),
		KeyPrefix:  prefix,
		Scopes:     []string{"kyc:initiate", "kyc:status"},
		RateLimit:  100,
		IsActive:   true,
	}
	err := repo.CreateAPIKey(ctx, key)
	require.NoError(t, err)

	found, err := repo.ValidateKey(ctx, rawKey)
	require.NoError(t, err)
	assert.Equal(t, "validate-test", found.ClientName)
	assert.Equal(t, []string{"kyc:initiate", "kyc:status"}, found.Scopes)
}

func TestRepository_ValidateKey_InvalidLength(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	_, err := repo.ValidateKey(ctx, "short")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid key length")
}

func TestRepository_ValidateKey_WrongPrefix(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	_, err := repo.ValidateKey(ctx, "xx_invalid_prefix_12345678")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid key prefix")
}

func TestRepository_ValidateKey_HashMismatch(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	rawKey := "dk_hash_mismatch_test_key"
	prefix := rawKey[:12]

	key := &model.APIKey{
		ClientName: "hash-mismatch",
		KeyHash:    "different_hash_value_here_does_not_match",
		KeyPrefix:  prefix,
		Scopes:     []string{"kyc:initiate"},
		IsActive:   true,
	}
	err := repo.CreateAPIKey(ctx, key)
	require.NoError(t, err)

	_, err = repo.ValidateKey(ctx, rawKey)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "key hash mismatch")
}

func TestRepository_UpdateLastUsed(t *testing.T) {
	repo, ctx := setupTestRepo(t)

	key := &model.APIKey{
		ClientName: "last-used-test",
		KeyHash:    "last_used_hash_value_12345",
		KeyPrefix:  "dk_last_used",
		Scopes:     []string{"kyc:initiate"},
		IsActive:   true,
	}
	err := repo.CreateAPIKey(ctx, key)
	require.NoError(t, err)

	err = repo.UpdateLastUsed(ctx, key.ID)
	require.NoError(t, err)

	found, err := repo.FindByPrefix(ctx, "dk_last_used")
	require.NoError(t, err)
	assert.NotNil(t, found.LastUsedAt)
}
