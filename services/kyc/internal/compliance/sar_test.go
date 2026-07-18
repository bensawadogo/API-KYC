package compliance_test

import (
	"context"
	"testing"

	"github.com/datakeys/kyc-service/internal/compliance"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func newSARGenerator(t *testing.T) *compliance.SARGenerator {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	return compliance.NewSARGenerator(logger)
}

func TestSARGenerator_Generate(t *testing.T) {
	g := newSARGenerator(t)
	req := compliance.SARRequest{
		VerificationID:  "v-12345",
		FullName:        "Mamadou Diop",
		Phone:           "+221771234567",
		CountryCode:     "SN",
		DocType:         "NATIONAL_ID",
		RiskScore:       0.85,
		Flags:           []string{"SANCTIONS_MATCH", "PEP_DETECTED"},
		AMLMatchDetails: "Matched against EU sanctions list entry #EU-2024-123",
		Regulation:      "BCEAO",
	}

	report, err := g.Generate(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, "v-12345", report.VerificationID)
	assert.Equal(t, "Mamadou Diop", report.EntityName)
	assert.Equal(t, "SN", report.CountryCode)
	assert.Equal(t, 0.85, report.RiskScore)
	assert.Equal(t, compliance.SARStatusDraft, report.FilingStatus)
	assert.Contains(t, report.Narrative, "BCEAO")
	assert.Contains(t, report.Narrative, "Mamadou Diop")
	assert.NotEmpty(t, report.CreatedAt)
}

func TestSARGenerator_Generate_LowRisk(t *testing.T) {
	g := newSARGenerator(t)
	req := compliance.SARRequest{
		VerificationID: "v-67890",
		FullName:       "Fatou Sy",
		CountryCode:    "BF",
		RiskScore:      0.3,
		Regulation:     "BCEAO",
	}

	report, err := g.Generate(context.Background(), req)
	assert.NoError(t, err)
	assert.Equal(t, "v-67890", report.VerificationID)
	assert.NotContains(t, report.Narrative, "exceeds")
}

func TestSARGenerator_FormatBCEAO(t *testing.T) {
	g := newSARGenerator(t)
	report := &compliance.SARReport{
		VerificationID: "v-12345",
		EntityName:     "Amadou Toure",
		CountryCode:    "ML",
		RiskScore:      0.9,
		Flags:          []string{"SANCTIONS_MATCH"},
		Narrative:      "High-risk match detected",
		Regulation:     "BCEAO",
	}

	output := g.FormatBCEAO(report)
	assert.Contains(t, output, "RAPPORT BCEAO")
	assert.Contains(t, output, "Amadou Toure")
	assert.Contains(t, output, "ML")
	assert.Contains(t, output, "Instruction N°008-09-2015")
	assert.Contains(t, output, "BROUILLON")
}

func TestSARGenerator_FormatEU(t *testing.T) {
	g := newSARGenerator(t)
	report := &compliance.SARReport{
		VerificationID: "v-67890",
		EntityName:     "John Smith",
		CountryCode:    "FR",
		RiskScore:      0.75,
		Flags:          []string{"PEP_DETECTED"},
		Narrative:      "PEP match requiring EDD",
		Regulation:     "AMLD",
	}

	output := g.FormatEU(report)
	assert.Contains(t, output, "EU AML DIRECTIVE 6")
	assert.Contains(t, output, "John Smith")
	assert.Contains(t, output, "FR")
	assert.Contains(t, output, "AMLD6")
	assert.Contains(t, output, "DRAFT")
}

func TestSARGenerator_GenerateWithAMLDetails(t *testing.T) {
	g := newSARGenerator(t)
	req := compliance.SARRequest{
		VerificationID:  "v-11111",
		FullName:        "Jean Claude",
		CountryCode:     "CI",
		RiskScore:       0.95,
		Flags:           []string{"SANCTIONS_MATCH", "PEP_DETECTED", "EDD_REQUIRED_HIGH_VALUE"},
		AMLMatchDetails: "OFAC SDN #12345, EU sanctions list #EU-2025-001, UNSC resolution 1718",
		Regulation:      "BCEAO",
	}

	report, err := g.Generate(context.Background(), req)
	assert.NoError(t, err)
	assert.Contains(t, report.Narrative, "OFAC")
	assert.Contains(t, report.Narrative, "EU-2025-001")
	assert.Len(t, report.Flags, 3)
}

func TestSARGenerator_BCEAOFormatIncludesNarrative(t *testing.T) {
	g := newSARGenerator(t)
	report := &compliance.SARReport{
		VerificationID: "v-55555",
		EntityName:     "Test User",
		CountryCode:    "BF",
		Narrative:      "Custom narrative for testing purposes",
		Regulation:     "BCEAO",
	}

	output := g.FormatBCEAO(report)
	assert.Contains(t, output, "Custom narrative for testing purposes")
}

func TestSARGenerator_EUFormatIncludesNarrative(t *testing.T) {
	g := newSARGenerator(t)
	report := &compliance.SARReport{
		VerificationID: "v-66666",
		EntityName:     "Test User EU",
		CountryCode:    "DE",
		Narrative:      "SAR narrative for EU filing",
		Regulation:     "AMLD",
	}

	output := g.FormatEU(report)
	assert.Contains(t, output, "SAR narrative for EU filing")
}
