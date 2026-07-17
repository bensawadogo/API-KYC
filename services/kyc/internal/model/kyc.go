package model

const (
	StatusPending      = "pending"
	StatusProcessing   = "processing"
	StatusApproved     = "approved"
	StatusRejected     = "rejected"
	StatusExpired      = "expired"
	StatusManualReview = "manual_review"
)

const (
	FlagExpiredDoc      = "EXPIRED_DOC"
	FlagSanctionsMatch  = "SANCTIONS_MATCH"
	FlagInvalidFormat   = "INVALID_FORMAT"
	FlagLowConfidence   = "LOW_CONFIDENCE"
	FlagDuplicatePhone  = "DUPLICATE_PHONE"
	FlagManualRequired  = "MANUAL_REVIEW_REQUIRED"
	FlagUnsupportedDoc  = "UNSUPPORTED_DOC_TYPE"
	FlagCountryMismatch = "COUNTRY_PHONE_MISMATCH"
	FlagPEPDetected         = "PEP_DETECTED"
	FlagAMLCheckFailed      = "AML_CHECK_FAILED"
	FlagProviderUnavailable = "PROVIDER_UNAVAILABLE"
)

type InitiateKYCRequest struct {
	Phone       string `json:"phone" validate:"required,e164"`
	CountryCode string `json:"country_code" validate:"required,len=2"`
	DocType     string `json:"doc_type" validate:"required"`
	DocNumber   string `json:"doc_number" validate:"omitempty"`
	FullName    string `json:"full_name" validate:"required,min=2,max=100"`
	Consent     bool   `json:"consent" validate:"required"`
	CallbackURL string `json:"callback_url" validate:"omitempty,url"`
	Language    string `json:"language" validate:"omitempty,oneof=fr en ar sw ha"`
}

type InitiateKYCResponse struct {
	VerificationID string `json:"verification_id"`
	Status         string `json:"status"`
	UploadURL      string `json:"upload_url"`
	ExpiresIn      int    `json:"expires_in"`
	Provider       string `json:"provider"`
	Message        string `json:"message"`
}

type VerificationResult struct {
	VerificationID string   `json:"verification_id"`
	Phone          string   `json:"phone"`
	CountryCode    string   `json:"country_code"`
	DocType        string   `json:"doc_type"`
	DocNumber      string   `json:"-"`
	Status         string   `json:"status"`
	Score          float64  `json:"score"`
	Provider       string   `json:"provider"`
	Flags          []string `json:"flags"`
	CallbackURL    string   `json:"-"`
	Consent        bool     `json:"-"`
	ProcessedAt    string   `json:"processed_at,omitempty"`
	ExpiresAt      string   `json:"expires_at"`
	AMLScore       float64  `json:"aml_score,omitempty"`
	IsSanctioned   bool     `json:"is_sanctioned"`
	IsPEP          bool     `json:"is_pep"`
}

type WebhookPayload struct {
	Event     string             `json:"event"`
	Data      VerificationResult `json:"data"`
	Signature string             `json:"signature"`
}
