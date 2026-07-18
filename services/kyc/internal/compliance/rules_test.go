package compliance_test

import (
	"context"
	"testing"

	"github.com/datakeys/kyc-service/internal/compliance"
	"github.com/datakeys/kyc-service/internal/countries"
	"github.com/datakeys/kyc-service/internal/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func newEvaluator(t *testing.T) *compliance.RuleEvaluator {
	t.Helper()
	logger, _ := zap.NewDevelopment()
	return compliance.NewRuleEvaluator(logger)
}

func TestBCEAOConsentRule_BlocksWithoutConsent(t *testing.T) {
	e := newEvaluator(t)
	country, _ := countries.GetCountry("BF")
	req := &compliance.RuleRequest{Consent: false}

	results, err := e.Evaluate(context.Background(), *country, req)
	assert.NoError(t, err)
	assert.True(t, e.HasBlocked(results), "BCEAO exige le consentement")
}

func TestBCEAOConsentRule_PassesWithConsent(t *testing.T) {
	e := newEvaluator(t)
	country, _ := countries.GetCountry("SN")
	req := &compliance.RuleRequest{Consent: true}

	results, err := e.Evaluate(context.Background(), *country, req)
	assert.NoError(t, err)
	assert.False(t, e.HasBlocked(results), "consentement donne")
}

func TestBCEAODataLocalization_PhoneMismatch(t *testing.T) {
	e := newEvaluator(t)
	country, _ := countries.GetCountry("CI")
	req := &compliance.RuleRequest{
		Phone:       "+33612345678",
		CountryCode: "CI",
		Consent:     true,
	}

	results, err := e.Evaluate(context.Background(), *country, req)
	assert.NoError(t, err)
	flags := e.MergeFlags(results)
	assert.Contains(t, flags, model.FlagCountryMismatch)
}

func TestBCEAODataLocalization_PhoneMatch(t *testing.T) {
	e := newEvaluator(t)
	country, _ := countries.GetCountry("CI")
	req := &compliance.RuleRequest{
		Phone:       "+22512345678",
		CountryCode: "CI",
		Consent:     true,
	}

	results, err := e.Evaluate(context.Background(), *country, req)
	assert.NoError(t, err)
	flags := e.MergeFlags(results)
	assert.NotContains(t, flags, model.FlagCountryMismatch)
}

func TestGDPRConsentRule_BlocksWithoutConsent(t *testing.T) {
	e := newEvaluator(t)
	country := countries.Country{
		Code: "FR", Name: "France", Regulations: []string{"GDPR", "AMLD"},
	}
	req := &compliance.RuleRequest{Consent: false}

	results, err := e.Evaluate(context.Background(), country, req)
	assert.NoError(t, err)
	assert.True(t, e.HasBlocked(results), "GDPR Art. 7 exige consentement")
}

func TestGDPRConsentRule_PassesWithConsent(t *testing.T) {
	e := newEvaluator(t)
	country := countries.Country{
		Code: "DE", Name: "Germany", Regulations: []string{"GDPR", "AMLD"},
	}
	req := &compliance.RuleRequest{Consent: true}

	results, err := e.Evaluate(context.Background(), country, req)
	assert.NoError(t, err)
	assert.False(t, e.HasBlocked(results), "consentement donne")
}

func TestGDPRDataMinimization_LongName(t *testing.T) {
	e := newEvaluator(t)
	country := countries.Country{
		Code: "FR", Name: "France", Regulations: []string{"GDPR", "AMLD"},
	}
	req := &compliance.RuleRequest{
		FullName: "Jean-Patrick-Christophe-Andre-Michel de la Fontaine des Saint-Peres du Jardin et de la Rochefoucauld de Bordeaux-Montesquieu",
		Consent:  true,
	}

	results, err := e.Evaluate(context.Background(), country, req)
	assert.NoError(t, err)
	flags := e.MergeFlags(results)
	assert.Contains(t, flags, model.FlagManualRequired)
}

func TestAMLDEDDRule_HighValueTransaction(t *testing.T) {
	e := newEvaluator(t)
	country := countries.Country{
		Code: "FR", Name: "France", Regulations: []string{"GDPR", "AMLD"},
	}
	req := &compliance.RuleRequest{
		FullName: "Moussa Diallo",
		Consent:  true,
		Amount:   25000,
	}

	results, err := e.Evaluate(context.Background(), country, req)
	assert.NoError(t, err)
	score := e.MaxRiskScore(results)
	assert.GreaterOrEqual(t, score, 0.8, "AMLD EDD recommende pour >10 000 EUR")
}

func TestAMLDThresholdRule_MandatoryVerification(t *testing.T) {
	e := newEvaluator(t)
	country := countries.Country{
		Code: "DE", Name: "Germany", Regulations: []string{"GDPR", "AMLD"},
	}
	req := &compliance.RuleRequest{
		Consent: true,
		Amount:  20000,
	}

	results, err := e.Evaluate(context.Background(), country, req)
	assert.NoError(t, err)
	flags := e.MergeFlags(results)
	assert.Contains(t, flags, "MANDATORY_VERIFICATION")
}

func TestECOWASPhoneRule_ValidPhone(t *testing.T) {
	e := newEvaluator(t)
	country, _ := countries.GetCountry("GH")
	req := &compliance.RuleRequest{
		Phone:       "+233501234567",
		CountryCode: "GH",
		Consent:     true,
	}

	results, err := e.Evaluate(context.Background(), *country, req)
	assert.NoError(t, err)
	flags := e.MergeFlags(results)
	assert.NotContains(t, flags, model.FlagCountryMismatch)
}

func TestECOWASPhoneRule_InvalidPhone(t *testing.T) {
	e := newEvaluator(t)
	country, _ := countries.GetCountry("GH")
	req := &compliance.RuleRequest{
		Phone:       "+33612345678",
		CountryCode: "GH",
		Consent:     true,
	}

	results, err := e.Evaluate(context.Background(), *country, req)
	assert.NoError(t, err)
	flags := e.MergeFlags(results)
	assert.Contains(t, flags, model.FlagCountryMismatch)
}

func TestNonUEMOACountry_SkipsBCEAORules(t *testing.T) {
	e := newEvaluator(t)
	country, _ := countries.GetCountry("ZA")
	req := &compliance.RuleRequest{Consent: false}

	results, err := e.Evaluate(context.Background(), *country, req)
	assert.NoError(t, err)
	assert.False(t, e.HasBlocked(results), "ZA n'est pas UEMOA, BCEAO ne s'applique pas")
}

func TestCountryIsEU_True(t *testing.T) {
	assert.True(t, compliance.CountryIsEU("FR"))
	assert.True(t, compliance.CountryIsEU("de"))
	assert.True(t, compliance.CountryIsEU("IT"))
}

func TestCountryIsEU_False(t *testing.T) {
	assert.False(t, compliance.CountryIsEU("BF"))
	assert.False(t, compliance.CountryIsEU("SN"))
	assert.False(t, compliance.CountryIsEU("US"))
}

func TestRuleEvaluator_MergeFlags_Dedup(t *testing.T) {
	e := newEvaluator(t)
	results := []*compliance.RuleResult{
		{Flags: []string{"A", "B"}},
		{Flags: []string{"B", "C"}},
	}
	flags := e.MergeFlags(results)
	assert.ElementsMatch(t, []string{"A", "B", "C"}, flags)
}

func TestRuleEvaluator_MaxRiskScore(t *testing.T) {
	e := newEvaluator(t)
	results := []*compliance.RuleResult{
		{RiskScore: 0.3},
		{RiskScore: 0.9},
		{RiskScore: 0.5},
	}
	assert.Equal(t, 0.9, e.MaxRiskScore(results))
}
