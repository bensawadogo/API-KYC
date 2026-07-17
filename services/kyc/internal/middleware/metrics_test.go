package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/v1/kyc/status/550e8400-e29b-41d4-a716-446655440000", "/v1/kyc/status/:id"},
		{"/v1/kyc/countries/CI", "/v1/kyc/countries/:code"},
		{"/v1/kyc/countries", "/v1/kyc/countries"},
		{"/health/live", "/health/live"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizePath(tt.input)
			assert.Equal(t, tt.expected, got)
		})
	}
}
