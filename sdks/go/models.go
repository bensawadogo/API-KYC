package datakeys

import "time"

type VerificationStatus string

const (
	StatusPending      VerificationStatus = "pending"
	StatusProcessing   VerificationStatus = "processing"
	StatusApproved     VerificationStatus = "approved"
	StatusRejected     VerificationStatus = "rejected"
	StatusManualReview VerificationStatus = "manual_review"
	StatusExpired      VerificationStatus = "expired"
)

type DocType string

const (
	DocNationalID      DocType = "NATIONAL_ID"
	DocPassport        DocType = "PASSPORT"
	DocDriversLicense  DocType = "DRIVERS_LICENSE"
	DocVoterCard       DocType = "VOTER_CARD"
	DocResidencePermit DocType = "RESIDENCE_PERMIT"
)

type KYCVerification struct {
	ID           string             `json:"id"`
	Object       string             `json:"object"`
	Livemode     bool               `json:"livemode"`
	Created      int64              `json:"created"`
	Status       VerificationStatus `json:"status"`
	PhoneHash    string             `json:"phone_hash"`
	CountryCode  string             `json:"country_code"`
	DocType      string             `json:"doc_type"`
	Score        float64            `json:"score"`
	Provider     string             `json:"provider"`
	Flags        []string           `json:"flags"`
	AMLScore     float64            `json:"aml_score"`
	IsSanctioned bool               `json:"is_sanctioned"`
	IsPEP        bool               `json:"is_pep"`
	UploadURL    string             `json:"upload_url,omitempty"`
	ExpiresIn    int                `json:"expires_in,omitempty"`
	ProcessedAt  *time.Time         `json:"processed_at,omitempty"`
}

func (v *KYCVerification) IsApproved() bool {
	return v.Status == StatusApproved
}

func (v *KYCVerification) IsRejected() bool {
	return v.Status == StatusRejected
}

func (v *KYCVerification) IsTerminal() bool {
	switch v.Status {
	case StatusApproved, StatusRejected, StatusManualReview, StatusExpired:
		return true
	}
	return false
}

type InitiateParams struct {
	Phone       string  `json:"phone"`
	CountryCode string  `json:"country_code"`
	DocType     DocType `json:"doc_type"`
	DocNumber   string  `json:"doc_number,omitempty"`
	FullName    string  `json:"full_name"`
	Consent     bool    `json:"consent"`
	CallbackURL string  `json:"callback_url,omitempty"`
	Language    string  `json:"language,omitempty"`
}

type apiResponse[T any] struct {
	Success   bool    `json:"success"`
	Data      *T      `json:"data"`
	Error     *string `json:"error"`
	Timestamp string  `json:"timestamp"`
}
