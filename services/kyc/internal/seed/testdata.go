package seed

import "fmt"

type TestProfile struct {
	Phone       string
	CountryCode string
	DocType     string
	DocNumber   string
	FullName    string
	Scenario    string
	Description string
}

var SandboxProfiles = []TestProfile{
	{
		Phone:       "+22670000001",
		CountryCode: "BF",
		DocType:     "NATIONAL_ID",
		DocNumber:   "B1230000",
		FullName:    "Aminata Ouédraogo",
		Scenario:    "approved",
		Description: "Vérification réussie — profil nominal BF",
	},
	{
		Phone:       "+2348100000001",
		CountryCode: "NG",
		DocType:     "NATIONAL_ID",
		DocNumber:   "12345670000",
		FullName:    "Chukwuemeka Okafor",
		Scenario:    "approved",
		Description: "Vérification réussie — NIN Nigeria",
	},
	{
		Phone:       "+221770004444",
		CountryCode: "SN",
		DocType:     "NATIONAL_ID",
		DocNumber:   "SN4444444",
		FullName:    "Fatou Diallo",
		Scenario:    "expired_doc",
		Description: "Document expiré — test cas d'erreur",
	},
	{
		Phone:       "+254700002222",
		CountryCode: "KE",
		DocType:     "NATIONAL_ID",
		DocNumber:   "KE2222222",
		FullName:    "John Mwangi",
		Scenario:    "sanctions",
		Description: "Match liste sanctions — test AML",
	},
	{
		Phone:       "+2250700003333",
		CountryCode: "CI",
		DocType:     "NATIONAL_ID",
		DocNumber:   "CI3333333",
		FullName:    "Kofi Asante",
		Scenario:    "pep",
		Description: "PEP détecté — test manual review",
	},
	{
		Phone:       "+212600000001",
		CountryCode: "MA",
		DocType:     "PASSPORT",
		DocNumber:   "MA0000001",
		FullName:    "Fatima Zahra Benali",
		Scenario:    "approved",
		Description: "Passeport Maroc — approuvé",
	},
	{
		Phone:       "+233200001111",
		CountryCode: "GH",
		DocType:     "NATIONAL_ID",
		DocNumber:   "GHA-1111111-1",
		FullName:    "Kwame Asare",
		Scenario:    "rejected",
		Description: "Score faible — test rejet",
	},
	{
		Phone:       "+27810000001",
		CountryCode: "ZA",
		DocType:     "NATIONAL_ID",
		DocNumber:   "9001010000087",
		FullName:    "Thabo Nkosi",
		Scenario:    "approved",
		Description: "ID Afrique du Sud — approuvé",
	},
}

func GetTestAPIKey() string {
	return "dk_test_datakeys_sandbox_001"
}

func CurlExample(profile TestProfile, baseURL, apiKey string) string {
	return fmt.Sprintf(`curl -X POST %s/v1/kyc/initiate \
  -H "X-API-Key: %s" \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "%s",
    "country_code": "%s",
    "doc_type": "%s",
    "doc_number": "%s",
    "full_name": "%s",
    "consent": true
  }'`,
		baseURL, apiKey,
		profile.Phone, profile.CountryCode, profile.DocType,
		profile.DocNumber, profile.FullName,
	)
}
