package compliance

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"
)

type SARStatus string

const (
	SARStatusDraft     SARStatus = "draft"
	SARStatusSubmitted SARStatus = "submitted"
	SARStatusFiled     SARStatus = "filed"
)

type SARReport struct {
	ID              string    `json:"id"`
	VerificationID  string    `json:"verification_id"`
	EntityName      string    `json:"entity_name"`
	Phone           string    `json:"-"`
	CountryCode     string    `json:"country_code"`
	DocType         string    `json:"doc_type"`
	RiskScore       float64   `json:"risk_score"`
	Flags           []string  `json:"flags"`
	AMLMatchDetails string    `json:"aml_match_details"`
	Regulation      string    `json:"regulation"`
	Narrative       string    `json:"narrative"`
	FilingStatus    SARStatus `json:"filing_status"`
	FiledAt         *time.Time `json:"filed_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type SARGenerator struct {
	logger *zap.Logger
}

func NewSARGenerator(logger *zap.Logger) *SARGenerator {
	return &SARGenerator{logger: logger}
}

func (g *SARGenerator) Generate(ctx context.Context, req SARRequest) (*SARReport, error) {
	narrative := g.buildNarrative(req)
	now := time.Now().UTC()

	report := &SARReport{
		VerificationID:  req.VerificationID,
		EntityName:      req.FullName,
		Phone:           req.Phone,
		CountryCode:     req.CountryCode,
		DocType:         req.DocType,
		RiskScore:       req.RiskScore,
		Flags:           req.Flags,
		AMLMatchDetails: req.AMLMatchDetails,
		Regulation:      req.Regulation,
		Narrative:       narrative,
		FilingStatus:    SARStatusDraft,
		CreatedAt:       now,
	}

	g.logger.Info("SAR report generated",
		zap.String("verification_id", req.VerificationID),
		zap.Float64("risk_score", req.RiskScore),
		zap.String("regulation", req.Regulation),
	)

	return report, nil
}

func (g *SARGenerator) FormatBCEAO(report *SARReport) string {
	var b strings.Builder
	b.WriteString("=== RAPPORT BCEAO D'OPERATION SUSPECTE ===\n")
	b.WriteString(fmt.Sprintf("Date: %s\n", report.CreatedAt.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("ID Rapport: %s\n", report.ID))
	b.WriteString(fmt.Sprintf("ID Verification: %s\n", report.VerificationID))
	b.WriteString(fmt.Sprintf("Entite: %s\n", report.EntityName))
	b.WriteString(fmt.Sprintf("Pays: %s\n", report.CountryCode))
	b.WriteString(fmt.Sprintf("Type Piece: %s\n", report.DocType))
	b.WriteString(fmt.Sprintf("Score Risque: %.0f%%\n", report.RiskScore*100))
	b.WriteString(fmt.Sprintf("Drapeaux: %s\n", strings.Join(report.Flags, ", ")))
	b.WriteString("\nNARRATIF:\n")
	b.WriteString(report.Narrative)
	b.WriteString("\n\nReference BCEAO: Instruction N°008-09-2015 relative aux LBC/FT\n")
	b.WriteString("Statut: BROUILLON (a soumettre a la cellule de renseignement financier)\n")
	return b.String()
}

func (g *SARGenerator) FormatEU(report *SARReport) string {
	var b strings.Builder
	b.WriteString("=== EU AML DIRECTIVE 6 - SUSPICIOUS ACTIVITY REPORT ===\n")
	b.WriteString(fmt.Sprintf("Date: %s\n", report.CreatedAt.Format(time.RFC3339)))
	b.WriteString(fmt.Sprintf("Report ID: %s\n", report.ID))
	b.WriteString(fmt.Sprintf("Verification ID: %s\n", report.VerificationID))
	b.WriteString(fmt.Sprintf("Entity: %s\n", report.EntityName))
	b.WriteString(fmt.Sprintf("Country: %s\n", report.CountryCode))
	b.WriteString(fmt.Sprintf("Document Type: %s\n", report.DocType))
	b.WriteString(fmt.Sprintf("Risk Score: %.0f%%\n", report.RiskScore*100))
	b.WriteString(fmt.Sprintf("Flags: %s\n", strings.Join(report.Flags, ", ")))
	b.WriteString("\nNARRATIVE:\n")
	b.WriteString(report.Narrative)
	b.WriteString("\n\nReference: Directive (EU) 2024/1640 (AMLD6)\n")
	b.WriteString("Status: DRAFT (to be filed with FIU)\n")
	return b.String()
}

func (g *SARGenerator) buildNarrative(req SARRequest) string {
	var parts []string

	parts = append(parts, fmt.Sprintf(
		"KYC verification %s for %s (%s, %s) triggered regulatory review.",
		req.VerificationID, req.FullName, req.CountryCode, req.DocType,
	))

	if req.RiskScore > 0.6 {
		parts = append(parts, fmt.Sprintf(
			"Risk score of %.0f%% exceeds standard threshold, indicating potential suspicious activity.",
			req.RiskScore*100,
		))
	}

	if len(req.Flags) > 0 {
		parts = append(parts, fmt.Sprintf("Flags detected: %s.", strings.Join(req.Flags, ", ")))
	}

	if req.AMLMatchDetails != "" {
		parts = append(parts, fmt.Sprintf("AML screening results: %s.", req.AMLMatchDetails))
	}

	parts = append(parts, fmt.Sprintf(
		"Regulation basis: %s. This report is generated automatically and requires review.",
		req.Regulation,
	))

	return strings.Join(parts, " ")
}

type SARRequest struct {
	VerificationID  string
	FullName        string
	Phone           string
	CountryCode     string
	DocType         string
	RiskScore       float64
	Flags           []string
	AMLMatchDetails string
	Regulation      string
}
