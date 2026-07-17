package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/datakeys/kyc-service/config"
	"github.com/datakeys/kyc-service/internal"
	"github.com/datakeys/kyc-service/internal/countries"
	"github.com/datakeys/kyc-service/internal/model"
	"github.com/datakeys/kyc-service/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

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

func validBFCountry() *countries.Country {
	return &countries.Country{
		Code: "BF", Name: "Burkina Faso", PhonePrefix: "+226",
		Region: "WEST_AFRICA", Provider: "smileid",
		DocTypes: []countries.DocType{
			{Code: "NATIONAL_ID", Name: "CNIB", Pattern: `^[A-Z]{1}[0-9]{7}$`},
			{Code: "PASSPORT", Name: "Passeport"},
		},
	}
}

func validBFRequest() *model.InitiateKYCRequest {
	return &model.InitiateKYCRequest{
		Phone: "+22670000000", CountryCode: "BF",
		DocType: "NATIONAL_ID", DocNumber: "B1234567",
		FullName: "Test User", Consent: true,
	}
}

func setupService(t *testing.T) (*service.KYCService, *MockKYCRepository, *MockDocumentStorage, *MockIdentityProvider, *MockWebhookSender, *MockCountryRegistry) {
	t.Helper()
	repo := new(MockKYCRepository)
	storage := new(MockDocumentStorage)
	provider := new(MockIdentityProvider)
	webhook := new(MockWebhookSender)
	registry := new(MockCountryRegistry)
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

	providers := map[string]internal.IdentityProvider{"smileid": provider}
	svc := service.NewKYCService(repo, storage, providers, webhook, registry, cfg, logger, nil, nil, nil, nil)
	return svc, repo, storage, provider, webhook, registry
}

func TestInitiate_Success(t *testing.T) {
	svc, repo, storage, _, _, registry := setupService(t)

	registry.On("GetCountry", "BF").Return(validBFCountry(), true)
	registry.On("IsDocTypeValid", "BF", "NATIONAL_ID").Return(true)
	registry.On("ValidateDocNumber", "BF", "NATIONAL_ID", "B1234567").Return(true)
	registry.On("GetProvider", "BF").Return("smileid")
	repo.On("ExistsApproved", mock.Anything, "+22670000000", "BF").Return(false, nil)
	storage.On("GenerateUploadURL", mock.Anything, mock.Anything, "NATIONAL_ID").Return("http://upload.url/doc", nil)
	repo.On("CreateVerification", mock.Anything, mock.AnythingOfType("*model.VerificationResult")).Return(nil)

	resp, err := svc.Initiate(context.Background(), validBFRequest())
	if err != nil {
		t.Fatal(err)
	}
	if resp.VerificationID == "" {
		t.Error("verification_id should not be empty")
	}
	if resp.Status != "pending" {
		t.Errorf("expected pending, got %s", resp.Status)
	}
	if resp.ExpiresIn <= 0 {
		t.Error("expires_in should be positive")
	}
}

func TestInitiate_ConsentRefused(t *testing.T) {
	svc, _, _, _, _, _ := setupService(t)
	req := validBFRequest()
	req.Consent = false

	_, err := svc.Initiate(context.Background(), req)
	if err == nil {
		t.Error("expected error for missing consent")
	}
}

func TestInitiate_UnsupportedCountry(t *testing.T) {
	svc, _, _, _, _, registry := setupService(t)
	registry.On("GetCountry", "XX").Return((*countries.Country)(nil), false)

	req := validBFRequest()
	req.CountryCode = "XX"
	_, err := svc.Initiate(context.Background(), req)
	if err == nil {
		t.Error("expected error for unsupported country")
	}
}

func TestInitiate_InvalidDocType(t *testing.T) {
	svc, _, _, _, _, registry := setupService(t)
	registry.On("GetCountry", "BF").Return(validBFCountry(), true)
	registry.On("IsDocTypeValid", "BF", "VOTER_CARD").Return(false)

	req := validBFRequest()
	req.DocType = "VOTER_CARD"
	_, err := svc.Initiate(context.Background(), req)
	if err == nil {
		t.Error("expected error for invalid doc type")
	}
}

func TestInitiate_InvalidDocNumber(t *testing.T) {
	svc, _, _, _, _, registry := setupService(t)
	registry.On("GetCountry", "BF").Return(validBFCountry(), true)
	registry.On("IsDocTypeValid", "BF", "NATIONAL_ID").Return(true)
	registry.On("ValidateDocNumber", "BF", "NATIONAL_ID", "BAD").Return(false)

	req := validBFRequest()
	req.DocNumber = "BAD"
	_, err := svc.Initiate(context.Background(), req)
	if err == nil {
		t.Error("expected error for invalid doc number")
	}
}

func TestInitiate_StorageError(t *testing.T) {
	svc, repo, storage, _, _, registry := setupService(t)
	registry.On("GetCountry", "BF").Return(validBFCountry(), true)
	registry.On("IsDocTypeValid", "BF", "NATIONAL_ID").Return(true)
	registry.On("ValidateDocNumber", "BF", "NATIONAL_ID", "B1234567").Return(true)
	registry.On("GetProvider", "BF").Return("smileid")
	repo.On("ExistsApproved", mock.Anything, "+22670000000", "BF").Return(false, nil)
	storage.On("GenerateUploadURL", mock.Anything, mock.Anything, "NATIONAL_ID").Return("", assert.AnError)

	_, err := svc.Initiate(context.Background(), validBFRequest())
	if err == nil {
		t.Error("expected error for storage failure")
	}
}

func TestInitiate_RepoError(t *testing.T) {
	svc, repo, storage, _, _, registry := setupService(t)
	registry.On("GetCountry", "BF").Return(validBFCountry(), true)
	registry.On("IsDocTypeValid", "BF", "NATIONAL_ID").Return(true)
	registry.On("ValidateDocNumber", "BF", "NATIONAL_ID", "B1234567").Return(true)
	registry.On("GetProvider", "BF").Return("smileid")
	repo.On("ExistsApproved", mock.Anything, "+22670000000", "BF").Return(false, nil)
	storage.On("GenerateUploadURL", mock.Anything, mock.Anything, "NATIONAL_ID").Return("http://url", nil)
	repo.On("CreateVerification", mock.Anything, mock.Anything).Return(assert.AnError)

	_, err := svc.Initiate(context.Background(), validBFRequest())
	if err == nil {
		t.Error("expected error for repo failure")
	}
}

func TestInitiate_DuplicateApproved_NonBlocking(t *testing.T) {
	svc, repo, storage, _, _, registry := setupService(t)
	registry.On("GetCountry", "BF").Return(validBFCountry(), true)
	registry.On("IsDocTypeValid", "BF", "NATIONAL_ID").Return(true)
	registry.On("ValidateDocNumber", "BF", "NATIONAL_ID", "B1234567").Return(true)
	registry.On("GetProvider", "BF").Return("smileid")
	repo.On("ExistsApproved", mock.Anything, "+22670000000", "BF").Return(true, nil)
	storage.On("GenerateUploadURL", mock.Anything, mock.Anything, "NATIONAL_ID").Return("http://url", nil)
	repo.On("CreateVerification", mock.Anything, mock.Anything).Return(nil)

	_, err := svc.Initiate(context.Background(), validBFRequest())
	if err != nil {
		t.Errorf("duplicate should be non-blocking, got: %v", err)
	}
}

func TestProcess_Approved(t *testing.T) {
	svc, repo, storage, provider, _, registry := setupService(t)
	registry.On("GetProvider", "BF").Return("smileid")

	verification := &model.VerificationResult{
		VerificationID: "v1", Phone: "+22670000000", CountryCode: "BF",
		DocType: "NATIONAL_ID", DocNumber: "B1234567",
		Status: "pending", Provider: "smileid",
	}
	repo.On("GetVerification", mock.Anything, "v1").Return(verification, nil)
	storage.On("GetDocument", mock.Anything, "v1").Return([]byte("fake-doc"), nil)
	repo.On("UpdateStatus", mock.Anything, "v1", "processing", 0.0, mock.Anything, "smileid").Return(nil)
	provider.On("Verify", mock.Anything, mock.AnythingOfType("internal.ProviderRequest")).Return(&internal.ProviderResult{
		Approved: true, Score: 0.95, Provider: "smileid", Flags: []string{},
	}, nil)
	repo.On("UpdateStatus", mock.Anything, "v1", "approved", 0.95, mock.Anything, "smileid").Return(nil)
	storage.On("DeleteDocument", mock.Anything, "v1").Return(nil)

	err := svc.Process(context.Background(), "v1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestProcess_Rejected_LowScore(t *testing.T) {
	svc, repo, storage, provider, _, _ := setupService(t)
	verification := &model.VerificationResult{
		VerificationID: "v1", CountryCode: "BF",
		DocType: "NATIONAL_ID", Status: "pending", Provider: "smileid",
	}
	repo.On("GetVerification", mock.Anything, "v1").Return(verification, nil)
	storage.On("GetDocument", mock.Anything, "v1").Return([]byte("doc"), nil)
	repo.On("UpdateStatus", mock.Anything, "v1", "processing", 0.0, mock.Anything, "smileid").Return(nil)
	provider.On("Verify", mock.Anything, mock.Anything).Return(&internal.ProviderResult{
		Approved: false, Score: 0.2, Flags: []string{"LOW_CONFIDENCE"}, Provider: "smileid",
	}, nil)
	repo.On("UpdateStatus", mock.Anything, "v1", "rejected", 0.2, []string{"LOW_CONFIDENCE"}, "smileid").Return(nil)

	err := svc.Process(context.Background(), "v1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestProcess_ManualReview_MidScore(t *testing.T) {
	svc, repo, storage, provider, _, _ := setupService(t)
	verification := &model.VerificationResult{
		VerificationID: "v1", CountryCode: "BF",
		DocType: "NATIONAL_ID", Status: "pending", Provider: "smileid",
	}
	repo.On("GetVerification", mock.Anything, "v1").Return(verification, nil)
	storage.On("GetDocument", mock.Anything, "v1").Return([]byte("doc"), nil)
	repo.On("UpdateStatus", mock.Anything, "v1", "processing", 0.0, mock.Anything, "smileid").Return(nil)
	provider.On("Verify", mock.Anything, mock.Anything).Return(&internal.ProviderResult{
		Approved: false, Score: 0.55, Provider: "smileid", Flags: []string{},
	}, nil)
	repo.On("UpdateStatus", mock.Anything, "v1", "manual_review", 0.55, mock.Anything, "smileid").Return(nil)

	err := svc.Process(context.Background(), "v1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestProcess_NotPending_Error(t *testing.T) {
	svc, repo, _, _, _, _ := setupService(t)
	verification := &model.VerificationResult{
		VerificationID: "v1", Status: "approved",
	}
	repo.On("GetVerification", mock.Anything, "v1").Return(verification, nil)

	err := svc.Process(context.Background(), "v1")
	if err == nil {
		t.Error("expected error for non-pending verification")
	}
}

func TestProcess_ProviderError(t *testing.T) {
	svc, repo, storage, provider, _, _ := setupService(t)
	verification := &model.VerificationResult{
		VerificationID: "v1", CountryCode: "BF",
		DocType: "NATIONAL_ID", Status: "pending", Provider: "smileid",
	}
	repo.On("GetVerification", mock.Anything, "v1").Return(verification, nil)
	storage.On("GetDocument", mock.Anything, "v1").Return([]byte("doc"), nil)
	repo.On("UpdateStatus", mock.Anything, "v1", "processing", 0.0, mock.Anything, "smileid").Return(nil)
	provider.On("Verify", mock.Anything, mock.Anything).Return((*internal.ProviderResult)(nil), assert.AnError)

	err := svc.Process(context.Background(), "v1")
	if err == nil {
		t.Error("expected error for provider failure")
	}
}

func TestProcess_WebhookFired(t *testing.T) {
	svc, repo, storage, provider, webhook, _ := setupService(t)
	verification := &model.VerificationResult{
		VerificationID: "v1", CountryCode: "BF",
		DocType: "NATIONAL_ID", Status: "pending", Provider: "smileid",
		CallbackURL: "https://client.com/webhook",
	}
	repo.On("GetVerification", mock.Anything, "v1").Return(verification, nil)
	storage.On("GetDocument", mock.Anything, "v1").Return([]byte("doc"), nil)
	repo.On("UpdateStatus", mock.Anything, "v1", "processing", 0.0, mock.Anything, "smileid").Return(nil)
	provider.On("Verify", mock.Anything, mock.Anything).Return(&internal.ProviderResult{
		Approved: true, Score: 0.95, Provider: "smileid",
	}, nil)
	repo.On("UpdateStatus", mock.Anything, "v1", "approved", 0.95, mock.Anything, "smileid").Return(nil)
	storage.On("DeleteDocument", mock.Anything, "v1").Return(nil)
	repo.On("GetVerification", mock.Anything, "v1").Return(verification, nil)
	webhook.On("Send", mock.Anything, "https://client.com/webhook", mock.AnythingOfType("*model.WebhookPayload")).Return(nil)

	err := svc.Process(context.Background(), "v1")
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(100 * time.Millisecond)
	webhook.AssertExpectations(t)
}