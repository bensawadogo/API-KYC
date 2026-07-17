package internal

import (
	"context"
	"time"

	"github.com/datakeys/kyc-service/internal/countries"
	"github.com/datakeys/kyc-service/internal/model"
)

type KYCRepository interface {
	CreateVerification(ctx context.Context, v *model.VerificationResult) error
	GetVerification(ctx context.Context, id string) (*model.VerificationResult, error)
	UpdateStatus(ctx context.Context, id, status string, score float64, flags []string, provider string) error
	ListByPhone(ctx context.Context, phone string, limit int) ([]*model.VerificationResult, error)
	ExistsApproved(ctx context.Context, phone, countryCode string) (bool, error)
}

type DocumentStorage interface {
	GenerateUploadURL(ctx context.Context, verificationID, docType string) (string, error)
	GetDocument(ctx context.Context, verificationID string) ([]byte, error)
	DeleteDocument(ctx context.Context, verificationID string) error
}

type IdentityProvider interface {
	Verify(ctx context.Context, req ProviderRequest) (*ProviderResult, error)
	SupportedCountries() []string
	Name() string
}

type WebhookSender interface {
	Send(ctx context.Context, url string, payload *model.WebhookPayload) error
}

type CountryRegistry interface {
	GetCountry(code string) (*countries.Country, bool)
	IsDocTypeValid(countryCode, docType string) bool
	ValidateDocNumber(countryCode, docType, number string) bool
	GetProvider(countryCode string) string
}

type KYCServiceInterface interface {
	Initiate(ctx context.Context, req *model.InitiateKYCRequest) (*model.InitiateKYCResponse, error)
	Process(ctx context.Context, verificationID string) error
	GetStatus(ctx context.Context, verificationID string) (*model.VerificationResult, error)
}

type ProviderRequest struct {
	VerificationID string
	CountryCode    string
	DocType        string
	DocNumber      string
	DocData        []byte
	Phone          string
}

type ProviderResult struct {
	Approved bool
	Score    float64
	Flags    []string
	Provider string
	RawData  map[string]interface{}
}

type AMLRequest struct {
	FullName    string
	CountryCode string
	DateOfBirth string
	DocNumber   string
}

type AMLResult struct {
	IsSanctioned bool
	IsPEP        bool
	Score        float64
	Matches      []AMLMatch
	ScreenedAt   time.Time
	Source       string
}

type AMLMatch struct {
	EntityName string
	EntityID   string
	Topics     []string
	Score      float64
	Dataset    string
}

type AMLChecker interface {
	Check(ctx context.Context, req AMLRequest) (*AMLResult, error)
	Name() string
}

type AMLRepository interface {
	SaveAMLResult(ctx context.Context, verificationID string, result *AMLResult) error
	GetAMLResult(ctx context.Context, verificationID string) (*AMLResult, error)
}

type APIKeyRepository interface {
	FindByPrefix(ctx context.Context, prefix string) (*model.APIKey, error)
	ValidateKey(ctx context.Context, rawKey string) (*model.APIKey, error)
	UpdateLastUsed(ctx context.Context, id string) error
}
