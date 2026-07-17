package service

import (
	"context"
	"testing"

	"github.com/datakeys/kyc-service/config"
	"github.com/datakeys/kyc-service/internal"
	"github.com/datakeys/kyc-service/internal/countries"
	"github.com/datakeys/kyc-service/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestResolveStatus_Approved(t *testing.T) {
	if s := resolveStatus(0.85, true, 0.70); s != model.StatusApproved {
		t.Errorf("expected approved, got %s", s)
	}
}

func TestResolveStatus_Approved_LowScoreButApproved(t *testing.T) {
	if s := resolveStatus(0.5, true, 0.70); s != model.StatusManualReview {
		t.Errorf("expected manual_review (approved but below threshold), got %s", s)
	}
}

func TestResolveStatus_Rejected(t *testing.T) {
	if s := resolveStatus(0.2, false, 0.70); s != model.StatusRejected {
		t.Errorf("expected rejected, got %s", s)
	}
}

func TestResolveStatus_ManualReview(t *testing.T) {
	if s := resolveStatus(0.55, false, 0.70); s != model.StatusManualReview {
		t.Errorf("expected manual_review, got %s", s)
	}
}

func TestResolveStatus_ManualReview_Boundary(t *testing.T) {
	if s := resolveStatus(0.39, false, 0.70); s != model.StatusRejected {
		t.Errorf("expected rejected at 0.39, got %s", s)
	}
	if s := resolveStatus(0.40, false, 0.70); s != model.StatusManualReview {
		t.Errorf("expected manual_review at 0.40, got %s", s)
	}
}

func TestMergeFlags_Empty(t *testing.T) {
	result := mergeFlags(nil, nil)
	if len(result) != 0 {
		t.Errorf("expected empty, got %v", result)
	}
}

func TestMergeFlags_Dedup(t *testing.T) {
	result := mergeFlags([]string{"A", "B"}, []string{"B", "C"})
	if len(result) != 3 {
		t.Errorf("expected 3 unique flags, got %d: %v", len(result), result)
	}
}

func TestMergeFlags_EmptyStringsIgnored(t *testing.T) {
	result := mergeFlags([]string{""}, []string{"A", ""})
	if len(result) != 1 || result[0] != "A" {
		t.Errorf("expected only A, got %v", result)
	}
}

func TestGetStatus_Success(t *testing.T) {
	repo := new(MockKYCRepository)
	logger, _ := zap.NewDevelopment()
	cfg := config.Config{KYC: config.KYCConfig{ScoreThreshold: 0.70}}
	svc := NewKYCService(repo, nil, nil, nil, nil, cfg, logger, nil, nil, nil, nil)

	expected := &model.VerificationResult{
		VerificationID: "v1",
		Status:         "approved",
	}
	repo.On("GetVerification", mock.Anything, "v1").Return(expected, nil)

	result, err := svc.GetStatus(context.Background(), "v1")
	if err != nil {
		t.Fatal(err)
	}
	if result.VerificationID != "v1" {
		t.Errorf("expected v1, got %s", result.VerificationID)
	}
	if result.Status != "approved" {
		t.Errorf("expected approved, got %s", result.Status)
	}
}

func TestSendWebhook_Success(t *testing.T) {
	repo := new(MockKYCRepository)
	wh := new(MockWebhookSender)
	logger, _ := zap.NewDevelopment()
	cfg := config.Config{KYC: config.KYCConfig{ScoreThreshold: 0.70}}
	svc := NewKYCService(repo, nil, nil, wh, nil, cfg, logger, nil, nil, nil, nil)

	verification := &model.VerificationResult{
		VerificationID: "v1",
		Status:         "approved",
		Provider:       "smileid",
		Score:          0.95,
	}

	repo.On("GetVerification", mock.Anything, "v1").Return(verification, nil)
	wh.On("Send", mock.Anything, "https://hook.example.com", mock.AnythingOfType("*model.WebhookPayload")).Return(nil)

	svc.sendWebhook("https://hook.example.com", "v1")

	wh.AssertExpectations(t)
}

func TestSendWebhook_RepoError(t *testing.T) {
	repo := new(MockKYCRepository)
	wh := new(MockWebhookSender)
	logger, _ := zap.NewDevelopment()
	cfg := config.Config{KYC: config.KYCConfig{ScoreThreshold: 0.70}}
	svc := NewKYCService(repo, nil, nil, wh, nil, cfg, logger, nil, nil, nil, nil)

	repo.On("GetVerification", mock.Anything, "v1").Return((*model.VerificationResult)(nil), assert.AnError)

	svc.sendWebhook("https://hook.example.com", "v1")

	wh.AssertNotCalled(t, "Send", mock.Anything, mock.Anything, mock.Anything)
}

func TestSendWebhook_SendError(t *testing.T) {
	repo := new(MockKYCRepository)
	wh := new(MockWebhookSender)
	logger, _ := zap.NewDevelopment()
	cfg := config.Config{KYC: config.KYCConfig{ScoreThreshold: 0.70}}
	svc := NewKYCService(repo, nil, nil, wh, nil, cfg, logger, nil, nil, nil, nil)

	verification := &model.VerificationResult{
		VerificationID: "v1",
		Status:         "approved",
	}
	repo.On("GetVerification", mock.Anything, "v1").Return(verification, nil)
	wh.On("Send", mock.Anything, "https://hook.example.com", mock.Anything).Return(assert.AnError)

	svc.sendWebhook("https://hook.example.com", "v1")

	wh.AssertExpectations(t)
}

func TestSelectProvider_Preferred(t *testing.T) {
	svc, _, _, _, _, _ := setupService(t)
	registry := new(MockCountryRegistry)
	registry.On("GetProvider", "BF").Return("smileid")
	svc.registry = registry

	// The mockProvider is already registered as "smileid"
	result := svc.selectProvider("BF")
	if result != "smileid" {
		t.Errorf("expected smileid, got %s", result)
	}
}

func TestSelectProvider_Fallback(t *testing.T) {
	svc, _, _, _, _, _ := setupService(t)
	registry := new(MockCountryRegistry)
	registry.On("GetProvider", "BF").Return("sumsub")
	svc.registry = registry

	// sumsub is not in the providers map (only "smileid" is), so fallback to smileid
	result := svc.selectProvider("BF")
	if result != "smileid" {
		t.Errorf("expected smileid (fallback), got %s", result)
	}
}

func TestSelectProvider_OnlyLocal(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := config.Config{KYC: config.KYCConfig{ScoreThreshold: 0.70}}
	repo := new(MockKYCRepository)
	localProvider := new(MockIdentityProvider)

	localProvider.On("Name").Return("local").Maybe()
	localProvider.On("SupportedCountries").Return([]string{}).Maybe()

	providers := map[string]internal.IdentityProvider{"local": localProvider}
	svc := NewKYCService(repo, nil, providers, nil, nil, cfg, logger, nil, nil, nil, nil)

	registry := new(MockCountryRegistry)
	registry.On("GetProvider", "ZZ").Return("")
	svc.registry = registry

	result := svc.selectProvider("ZZ")
	if result != "local" {
		t.Errorf("expected local, got %s", result)
	}
}

func TestSelectProvider_WithDefault(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := config.Config{
		KYC: config.KYCConfig{
			ScoreThreshold:  0.70,
			DefaultProvider: "youverify",
		},
	}
	repo := new(MockKYCRepository)
	smileidProvider := new(MockIdentityProvider)
	youverifyProvider := new(MockIdentityProvider)

	smileidProvider.On("Name").Return("smileid").Maybe()
	smileidProvider.On("SupportedCountries").Return([]string{}).Maybe()
	youverifyProvider.On("Name").Return("youverify").Maybe()
	youverifyProvider.On("SupportedCountries").Return([]string{}).Maybe()

	providers := map[string]internal.IdentityProvider{
		"smileid":   smileidProvider,
		"youverify": youverifyProvider,
	}
	svc := NewKYCService(repo, nil, providers, nil, nil, cfg, logger, nil, nil, nil, nil)

	registry := new(MockCountryRegistry)
	registry.On("GetProvider", "XX").Return("")
	svc.registry = registry

	result := svc.selectProvider("XX")
	if result != "youverify" {
		t.Errorf("expected youverify (default), got %s", result)
	}
}

type MockKYCRepository struct{ mock.Mock }

func (m *MockKYCRepository) CreateVerification(ctx context.Context, v *model.VerificationResult) error {
	args := m.Called(ctx, v)
	return args.Error(0)
}

func (m *MockKYCRepository) GetVerification(ctx context.Context, id string) (*model.VerificationResult, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.VerificationResult), args.Error(1)
}

func (m *MockKYCRepository) UpdateStatus(ctx context.Context, id, status string, score float64, flags []string, provider string) error {
	args := m.Called(ctx, id, status, score, flags, provider)
	return args.Error(0)
}

func (m *MockKYCRepository) ListByPhone(ctx context.Context, phone string, limit int) ([]*model.VerificationResult, error) {
	args := m.Called(ctx, phone, limit)
	return args.Get(0).([]*model.VerificationResult), args.Error(1)
}

func (m *MockKYCRepository) ExistsApproved(ctx context.Context, phone, countryCode string) (bool, error) {
	args := m.Called(ctx, phone, countryCode)
	return args.Bool(0), args.Error(1)
}

type MockDocumentStorage struct{ mock.Mock }

func (m *MockDocumentStorage) GenerateUploadURL(ctx context.Context, verificationID, docType string) (string, error) {
	args := m.Called(ctx, verificationID, docType)
	return args.String(0), args.Error(1)
}

func (m *MockDocumentStorage) GetDocument(ctx context.Context, verificationID string) ([]byte, error) {
	args := m.Called(ctx, verificationID)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockDocumentStorage) DeleteDocument(ctx context.Context, verificationID string) error {
	args := m.Called(ctx, verificationID)
	return args.Error(0)
}

type MockIdentityProvider struct{ mock.Mock }

func (m *MockIdentityProvider) Verify(ctx context.Context, req internal.ProviderRequest) (*internal.ProviderResult, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*internal.ProviderResult), args.Error(1)
}

func (m *MockIdentityProvider) SupportedCountries() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func (m *MockIdentityProvider) Name() string {
	args := m.Called()
	return args.String(0)
}

type MockWebhookSender struct{ mock.Mock }

func (m *MockWebhookSender) Send(ctx context.Context, url string, payload *model.WebhookPayload) error {
	args := m.Called(ctx, url, payload)
	return args.Error(0)
}

type MockCountryRegistry struct{ mock.Mock }

func (m *MockCountryRegistry) GetCountry(code string) (*countries.Country, bool) {
	args := m.Called(code)
	return args.Get(0).(*countries.Country), args.Bool(1)
}

func (m *MockCountryRegistry) IsDocTypeValid(countryCode, docType string) bool {
	args := m.Called(countryCode, docType)
	return args.Bool(0)
}

func (m *MockCountryRegistry) ValidateDocNumber(countryCode, docType, number string) bool {
	args := m.Called(countryCode, docType, number)
	return args.Bool(0)
}

func (m *MockCountryRegistry) GetProvider(countryCode string) string {
	args := m.Called(countryCode)
	return args.String(0)
}

func setupService(t *testing.T) (*KYCService, *MockKYCRepository, *MockDocumentStorage, *MockIdentityProvider, *MockWebhookSender, *MockCountryRegistry) {
	t.Helper()

	repo := new(MockKYCRepository)
	docStorage := new(MockDocumentStorage)
	prov := new(MockIdentityProvider)
	wh := new(MockWebhookSender)
	reg := new(MockCountryRegistry)

	logger, _ := zap.NewDevelopment()
	cfg := config.Config{
		KYC: config.KYCConfig{
			MaxDocSizeMB: 10, SessionTTLSeconds: 3600,
			ScoreThreshold: 0.70, DefaultProvider: "smileid",
		},
		Compliance: config.ComplianceConfig{
			ConsentRequired: true, RetentionDays: 1825, AuditEnabled: false,
		},
	}
	providers := map[string]internal.IdentityProvider{"smileid": prov}
	svc := NewKYCService(repo, docStorage, providers, wh, reg, cfg, logger, nil, nil, nil, nil)
	return svc, repo, docStorage, prov, wh, reg
}
