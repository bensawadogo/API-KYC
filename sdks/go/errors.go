package datakeys

import "fmt"

type ErrorCode string

const (
	ErrAuthMissing ErrorCode = "KYC_AUTH_001"
	ErrAuthInvalid ErrorCode = "KYC_AUTH_002"
	ErrRateLimit   ErrorCode = "KYC_RATE_001"
	ErrConsent     ErrorCode = "KYC_VAL_001"
	ErrSanctions   ErrorCode = "KYC_AML_SANCTION"
	ErrIdempotent  ErrorCode = "KYC_IDMP_001"
	ErrServerError ErrorCode = "KYC_SERVER_ERR"
	ErrNetwork     ErrorCode = "KYC_NETWORK"
	ErrTimeout     ErrorCode = "KYC_TIMEOUT"
	ErrUnknown     ErrorCode = "KYC_UNKNOWN"
)

type KYCError struct {
	Code    ErrorCode
	Message string
	Status  int
	Raw     map[string]any
}

func (e *KYCError) Error() string {
	return fmt.Sprintf("[%s] %s (HTTP %d)", e.Code, e.Message, e.Status)
}

func (e *KYCError) IsAuthError() bool {
	return len(e.Code) >= 8 && e.Code[:8] == "KYC_AUTH"
}

func (e *KYCError) IsRateLimit() bool {
	return e.Code == ErrRateLimit
}

func (e *KYCError) IsSanctions() bool {
	return e.Code == ErrSanctions
}

func (e *KYCError) IsServerError() bool {
	return e.Status >= 500
}
