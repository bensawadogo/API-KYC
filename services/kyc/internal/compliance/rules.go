package compliance

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/datakeys/kyc-service/internal/countries"
	"github.com/datakeys/kyc-service/internal/model"
	"go.uber.org/zap"
)

type Regulation string

const (
	RegBCEAO  Regulation = "BCEAO"
	RegUEMOA  Regulation = "UEMOA"
	RegECOWAS Regulation = "ECOWAS"
	RegGDPR   Regulation = "GDPR"
	RegAMLD   Regulation = "AMLD" // EU Anti-Money Laundering Directive
	RegCBN    Regulation = "CBN"
	RegNDPR   Regulation = "NDPR"
	RegPOPIA  Regulation = "POPIA"
	RegBEAC   Regulation = "BEAC"
	RegCEMAC  Regulation = "CEMAC"
	RegSARB   Regulation = "SARB"
	RegBAM    Regulation = "BAM"
	RegCNDP   Regulation = "CNDP"
	RegBOG    Regulation = "BOG"
)

var EUCountries = map[string]bool{
	"AT": true, "BE": true, "BG": true, "HR": true, "CY": true, "CZ": true,
	"DK": true, "EE": true, "FI": true, "FR": true, "DE": true, "GR": true,
	"HU": true, "IE": true, "IT": true, "LV": true, "LT": true, "LU": true,
	"MT": true, "NL": true, "PL": true, "PT": true, "RO": true, "SK": true,
	"SI": true, "ES": true, "SE": true,
}

type RuleResult struct {
	Blocked    bool
	Flags      []string
	RiskScore  float64
	Message    string
	Regulation Regulation
}

type Rule interface {
	Name() string
	Evaluate(ctx context.Context, req *RuleRequest) (*RuleResult, error)
	AppliesTo(countryCode string, regulations []string) bool
}

type RuleRequest struct {
	Phone       string
	CountryCode string
	DocType     string
	DocNumber   string
	FullName    string
	Consent     bool
	Amount      float64 // montant de la transaction le cas échéant
}

type RuleEvaluator struct {
	rules  []Rule
	logger *zap.Logger
}

func NewRuleEvaluator(logger *zap.Logger) *RuleEvaluator {
	rules := []Rule{
		&BCEAOConsentRule{},
		&BCEAODataLocalizationRule{},
		&GDPRConsentRule{},
		&GDPRDataMinimizationRule{},
		&AMLDEDDRule{},
		&AMLDThresholdRule{},
		&ECOWASPhoneRule{},
	}
	return &RuleEvaluator{rules: rules, logger: logger}
}

func (e *RuleEvaluator) Evaluate(ctx context.Context, country countries.Country, req *RuleRequest) ([]*RuleResult, error) {
	var results []*RuleResult
	for _, rule := range e.rules {
		if !rule.AppliesTo(country.Code, country.Regulations) {
			continue
		}
		result, err := rule.Evaluate(ctx, req)
		if err != nil {
			e.logger.Warn("rule evaluation error", zap.String("rule", rule.Name()), zap.Error(err))
			continue
		}
		results = append(results, result)
	}
	return results, nil
}

func (e *RuleEvaluator) Rules() []Rule { return e.rules }

type BCEAOConsentRule struct{}

func (r *BCEAOConsentRule) Name() string { return "BCEAO Consent" }

func (r *BCEAOConsentRule) AppliesTo(code string, regs []string) bool {
	return hasRegulation(regs, RegBCEAO, RegUEMOA)
}

func (r *BCEAOConsentRule) Evaluate(ctx context.Context, req *RuleRequest) (*RuleResult, error) {
	if !req.Consent {
		return &RuleResult{
			Blocked:    true,
			Flags:      []string{model.FlagManualRequired},
			RiskScore:  1.0,
			Message:    "BCEAO Instruction N°008-09-2015 exige le consentement explicite du client pour le traitement des données personnelles",
			Regulation: RegBCEAO,
		}, nil
	}
	return &RuleResult{Regulation: RegBCEAO}, nil
}

type BCEAODataLocalizationRule struct{}

func (r *BCEAODataLocalizationRule) Name() string { return "BCEAO Data Localization" }

func (r *BCEAODataLocalizationRule) AppliesTo(code string, regs []string) bool {
	return hasRegulation(regs, RegBCEAO, RegUEMOA)
}

func (r *BCEAODataLocalizationRule) Evaluate(ctx context.Context, req *RuleRequest) (*RuleResult, error) {
	phonePrefix := countries.GetPhonePrefix(req.CountryCode)
	if phonePrefix == "" {
		return &RuleResult{Regulation: RegBCEAO}, nil
	}
	if !strings.HasPrefix(req.Phone, phonePrefix) {
		return &RuleResult{
			Flags:      []string{model.FlagCountryMismatch},
			RiskScore:  0.6,
			Message:    "BCEAO exige que le numéro de téléphone corresponde au pays UEMOA déclaré",
			Regulation: RegBCEAO,
		}, nil
	}
	return &RuleResult{Regulation: RegBCEAO}, nil
}

type GDPRConsentRule struct{}

func (r *GDPRConsentRule) Name() string { return "GDPR Consent" }

func (r *GDPRConsentRule) AppliesTo(code string, regs []string) bool {
	return EUCountries[code] || hasRegulation(regs, RegGDPR)
}

func (r *GDPRConsentRule) Evaluate(ctx context.Context, req *RuleRequest) (*RuleResult, error) {
	if !req.Consent {
		return &RuleResult{
			Blocked:    true,
			Flags:      []string{model.FlagManualRequired},
			RiskScore:  1.0,
			Message:    "GDPR Art. 7 exige le consentement explicite avant traitement des données biométriques",
			Regulation: RegGDPR,
		}, nil
	}
	return &RuleResult{Regulation: RegGDPR}, nil
}

type GDPRDataMinimizationRule struct{}

func (r *GDPRDataMinimizationRule) Name() string { return "GDPR Data Minimization" }

func (r *GDPRDataMinimizationRule) AppliesTo(code string, regs []string) bool {
	return EUCountries[code] || hasRegulation(regs, RegGDPR)
}

func (r *GDPRDataMinimizationRule) Evaluate(ctx context.Context, req *RuleRequest) (*RuleResult, error) {
	if len(req.FullName) > 100 {
		return &RuleResult{
			Flags:      []string{model.FlagManualRequired},
			RiskScore:  0.3,
			Message:    "GDPR Art. 5(1)(c): les données doivent être limitées à ce qui est nécessaire",
			Regulation: RegGDPR,
		}, nil
	}
	return &RuleResult{Regulation: RegGDPR}, nil
}

type AMLDEDDRule struct{}

func (r *AMLDEDDRule) Name() string { return "AMLD Enhanced Due Diligence" }

func (r *AMLDEDDRule) AppliesTo(code string, regs []string) bool {
	return EUCountries[code] || hasRegulation(regs, RegAMLD)
}

func (r *AMLDEDDRule) Evaluate(ctx context.Context, req *RuleRequest) (*RuleResult, error) {
	var flags []string
	var riskScore float64

	if len(req.FullName) < 2 {
		riskScore = 0.7
		flags = append(flags, "EDD_REQUIRED_INCOMPLETE_NAME")
	}

	if req.Amount > 10000 {
		riskScore = max(riskScore, 0.8)
		flags = append(flags, "EDD_REQUIRED_HIGH_VALUE")
	}

	result := &RuleResult{
		Flags:      flags,
		RiskScore:  riskScore,
		Message:    "",
		Regulation: RegAMLD,
	}
	if riskScore > 0.6 {
		result.Message = fmt.Sprintf("AMLD6 Art. 18: Enhanced Due Diligence recommended (risk score: %.0f%%)", riskScore*100)
	}
	return result, nil
}

type AMLDThresholdRule struct{}

func (r *AMLDThresholdRule) Name() string { return "AMLD Threshold" }

func (r *AMLDThresholdRule) AppliesTo(code string, regs []string) bool {
	return EUCountries[code] || hasRegulation(regs, RegAMLD)
}

func (r *AMLDThresholdRule) Evaluate(ctx context.Context, req *RuleRequest) (*RuleResult, error) {
	if req.Amount > 15000 {
		return &RuleResult{
			Flags:      []string{"MANDATORY_VERIFICATION"},
			RiskScore:  0.9,
			Message:    "AMLD6 Art. 11: transaction >15 000 EUR requires mandatory identity verification",
			Regulation: RegAMLD,
		}, nil
	}
	return &RuleResult{Regulation: RegAMLD}, nil
}

type ECOWASPhoneRule struct{}

func (r *ECOWASPhoneRule) Name() string { return "ECOWAS Phone Validation" }

func (r *ECOWASPhoneRule) AppliesTo(code string, regs []string) bool {
	return hasRegulation(regs, RegECOWAS)
}

func (r *ECOWASPhoneRule) Evaluate(ctx context.Context, req *RuleRequest) (*RuleResult, error) {
	prefix := countries.GetPhonePrefix(req.CountryCode)
	if prefix == "" {
		return &RuleResult{Regulation: RegECOWAS}, nil
	}
	if !strings.HasPrefix(req.Phone, prefix) {
		return &RuleResult{
			Flags:      []string{model.FlagCountryMismatch},
			RiskScore:  0.5,
			Message:    "ECOWAS Supplementary Act A/SA.1/01/15: phone number must match country of residence",
			Regulation: RegECOWAS,
		}, nil
	}
	return &RuleResult{Regulation: RegECOWAS}, nil
}

type ConsentRecord struct {
	ID             string    `json:"id"`
	Phone          string    `json:"-"`
	PhoneHash      string    `json:"phone_hash"`
	EventType      string    `json:"event_type"` // "granted", "withdrawn", "renewed"
	ConsentVersion string    `json:"consent_version"`
	IPAddress      string    `json:"ip_address"`
	CreatedAt      time.Time `json:"created_at"`
}

type ConsentManager interface {
	RecordConsent(ctx context.Context, phone string, version string, ipAddress string) error
	RecordWithdrawal(ctx context.Context, phone string, ip string) error
	HasActiveConsent(ctx context.Context, phone string) (bool, error)
	GetConsentHistory(ctx context.Context, phone string) ([]ConsentRecord, error)
}

func hasRegulation(regs []string, targets ...Regulation) bool {
	for _, reg := range regs {
		for _, t := range targets {
			if strings.EqualFold(reg, string(t)) {
				return true
			}
		}
	}
	return false
}

func (e *RuleEvaluator) HasBlocked(results []*RuleResult) bool {
	for _, r := range results {
		if r.Blocked {
			return true
		}
	}
	return false
}

func (e *RuleEvaluator) MergeFlags(results []*RuleResult) []string {
	seen := map[string]bool{}
	var flags []string
	for _, r := range results {
		for _, f := range r.Flags {
			if !seen[f] {
				seen[f] = true
				flags = append(flags, f)
			}
		}
	}
	return flags
}

func (e *RuleEvaluator) MaxRiskScore(results []*RuleResult) float64 {
	var max float64
	for _, r := range results {
		if r.RiskScore > max {
			max = r.RiskScore
		}
	}
	return max
}

func CountryIsEU(code string) bool {
	return EUCountries[strings.ToUpper(code)]
}
