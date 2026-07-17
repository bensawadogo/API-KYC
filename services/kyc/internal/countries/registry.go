package countries

import (
	"regexp"
	"strings"
)

const (
	RegionWestAfrica   = "WEST_AFRICA"
	RegionNorthAfrica  = "NORTH_AFRICA"
	RegionEastAfrica   = "EAST_AFRICA"
	RegionCentralAfrica = "CENTRAL_AFRICA"
	RegionSouthernAfrica = "SOUTHERN_AFRICA"
)

type DocType struct {
	Code    string
	Name    string
	Pattern string
}

type Country struct {
	Code        string
	Name        string
	PhonePrefix string
	Region      string
	DocTypes    []DocType
	Provider    string
	Regulations []string
}

var registry = buildRegistry()

func buildRegistry() map[string]Country {
	countries := []Country{
		// Afrique de l'Ouest
		{
			Code: "BF", Name: "Burkina Faso", PhonePrefix: "+226", Region: RegionWestAfrica,
			Provider: "smileid", Regulations: []string{"BCEAO", "UEMOA"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNIB", Pattern: `^[A-Z]{1}[0-9]{7}$`},
				{Code: "PASSPORT", Name: "Passeport"},
				{Code: "DRIVERS_LICENSE", Name: "Permis de conduire"},
			},
		},
		{
			Code: "SN", Name: "Sénégal", PhonePrefix: "+221", Region: RegionWestAfrica,
			Provider: "smileid", Regulations: []string{"BCEAO", "UEMOA", "ECOWAS"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNI"},
				{Code: "PASSPORT", Name: "Passeport"},
				{Code: "DRIVERS_LICENSE", Name: "Permis de conduire"},
			},
		},
		{
			Code: "ML", Name: "Mali", PhonePrefix: "+223", Region: RegionWestAfrica,
			Provider: "smileid", Regulations: []string{"BCEAO", "UEMOA"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "NINA"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "CI", Name: "Côte d'Ivoire", PhonePrefix: "+225", Region: RegionWestAfrica,
			Provider: "smileid", Regulations: []string{"BCEAO", "UEMOA", "ECOWAS"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNI"},
				{Code: "PASSPORT", Name: "Passeport"},
				{Code: "DRIVERS_LICENSE", Name: "Permis de conduire"},
			},
		},
		{
			Code: "GH", Name: "Ghana", PhonePrefix: "+233", Region: RegionWestAfrica,
			Provider: "youverify", Regulations: []string{"ECOWAS", "BOG"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "Ghana Card", Pattern: `^GHA-[0-9]{9}-[0-9]$`},
				{Code: "PASSPORT", Name: "Passeport"},
				{Code: "VOTER_CARD", Name: "Voter ID"},
			},
		},
		{
			Code: "NG", Name: "Nigeria", PhonePrefix: "+234", Region: RegionWestAfrica,
			Provider: "youverify", Regulations: []string{"ECOWAS", "CBN", "NDPR"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "NIN", Pattern: `^[0-9]{11}$`},
				{Code: "BVN", Name: "Bank Verification Number", Pattern: `^[0-9]{11}$`},
				{Code: "PASSPORT", Name: "Passeport"},
				{Code: "VOTER_CARD", Name: "Voter Card"},
			},
		},
		{
			Code: "SL", Name: "Sierra Leone", PhonePrefix: "+232", Region: RegionWestAfrica,
			Provider: "smileid", Regulations: []string{"ECOWAS", "BSL"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNI"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "GN", Name: "Guinée", PhonePrefix: "+224", Region: RegionWestAfrica,
			Provider: "smileid", Regulations: []string{"ECOWAS", "BCRG"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNI"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "TG", Name: "Togo", PhonePrefix: "+228", Region: RegionWestAfrica,
			Provider: "smileid", Regulations: []string{"BCEAO", "UEMOA", "ECOWAS"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNI"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "BJ", Name: "Bénin", PhonePrefix: "+229", Region: RegionWestAfrica,
			Provider: "smileid", Regulations: []string{"BCEAO", "UEMOA", "ECOWAS"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNI"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "NE", Name: "Niger", PhonePrefix: "+227", Region: RegionWestAfrica,
			Provider: "smileid", Regulations: []string{"BCEAO", "UEMOA", "ECOWAS"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNI"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "GW", Name: "Guinée-Bissau", PhonePrefix: "+245", Region: RegionWestAfrica,
			Provider: "smileid", Regulations: []string{"ECOWAS", "BCEAO"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNI"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "LR", Name: "Libéria", PhonePrefix: "+231", Region: RegionWestAfrica,
			Provider: "smileid", Regulations: []string{"ECOWAS", "CBL"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNI"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "GM", Name: "Gambie", PhonePrefix: "+220", Region: RegionWestAfrica,
			Provider: "smileid", Regulations: []string{"ECOWAS"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNI"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "MR", Name: "Mauritanie", PhonePrefix: "+222", Region: RegionWestAfrica,
			Provider: "smileid", Regulations: []string{"ECOWAS"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNI"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "CV", Name: "Cap-Vert", PhonePrefix: "+238", Region: RegionWestAfrica,
			Provider: "smileid", Regulations: []string{"ECOWAS"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNI"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},

		// Afrique de l'Est
		{
			Code: "KE", Name: "Kenya", PhonePrefix: "+254", Region: RegionEastAfrica,
			Provider: "smileid", Regulations: []string{"EAC", "CBK"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "National ID", Pattern: `^[0-9]{8}$`},
				{Code: "PASSPORT", Name: "Passeport"},
				{Code: "KRA_PIN", Name: "KRA PIN", Pattern: `^[A-Z][0-9]{9}[A-Z]$`},
			},
		},
		{
			Code: "TZ", Name: "Tanzanie", PhonePrefix: "+255", Region: RegionEastAfrica,
			Provider: "smileid", Regulations: []string{"EAC", "BOT"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "NIDA"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "ET", Name: "Éthiopie", PhonePrefix: "+251", Region: RegionEastAfrica,
			Provider: "smileid", Regulations: []string{"NBE"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "Kebele ID"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "UG", Name: "Ouganda", PhonePrefix: "+256", Region: RegionEastAfrica,
			Provider: "smileid", Regulations: []string{"EAC", "BOU"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "National ID"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "RW", Name: "Rwanda", PhonePrefix: "+250", Region: RegionEastAfrica,
			Provider: "smileid", Regulations: []string{"EAC", "BNR"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "National ID"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "SS", Name: "Soudan du Sud", PhonePrefix: "+211", Region: RegionEastAfrica,
			Provider: "smileid", Regulations: []string{"EAC"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "National ID"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "SO", Name: "Somalie", PhonePrefix: "+252", Region: RegionEastAfrica,
			Provider: "smileid", Regulations: []string{},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "National ID"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "DJ", Name: "Djibouti", PhonePrefix: "+253", Region: RegionEastAfrica,
			Provider: "smileid", Regulations: []string{},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNI"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "ER", Name: "Érythrée", PhonePrefix: "+291", Region: RegionEastAfrica,
			Provider: "smileid", Regulations: []string{},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "National ID"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "BI", Name: "Burundi", PhonePrefix: "+257", Region: RegionEastAfrica,
			Provider: "smileid", Regulations: []string{"EAC"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNI"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "SC", Name: "Seychelles", PhonePrefix: "+248", Region: RegionEastAfrica,
			Provider: "smileid", Regulations: []string{},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "National ID"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "KM", Name: "Comores", PhonePrefix: "+269", Region: RegionEastAfrica,
			Provider: "smileid", Regulations: []string{},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNI"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "MG", Name: "Madagascar", PhonePrefix: "+261", Region: RegionEastAfrica,
			Provider: "smileid", Regulations: []string{"EAC"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNI"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "MU", Name: "Maurice", PhonePrefix: "+230", Region: RegionEastAfrica,
			Provider: "smileid", Regulations: []string{},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "National ID"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},

		// Afrique du Nord
		{
			Code: "MA", Name: "Maroc", PhonePrefix: "+212", Region: RegionNorthAfrica,
			Provider: "sumsub", Regulations: []string{"BAM", "CNDP"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CIN", Pattern: `^[A-Z]{1,2}[0-9]{5,6}$`},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "EG", Name: "Égypte", PhonePrefix: "+20", Region: RegionNorthAfrica,
			Provider: "sumsub", Regulations: []string{"CBE"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "National ID", Pattern: `^[0-9]{14}$`},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "TN", Name: "Tunisie", PhonePrefix: "+216", Region: RegionNorthAfrica,
			Provider: "sumsub", Regulations: []string{"BCT", "INPDP"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CIN", Pattern: `^[0-9]{8}$`},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "DZ", Name: "Algérie", PhonePrefix: "+213", Region: RegionNorthAfrica,
			Provider: "sumsub", Regulations: []string{"BA", "ANPDP"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNI"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "LY", Name: "Libye", PhonePrefix: "+218", Region: RegionNorthAfrica,
			Provider: "sumsub", Regulations: []string{"CBL"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNI"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "SD", Name: "Soudan", PhonePrefix: "+249", Region: RegionNorthAfrica,
			Provider: "sumsub", Regulations: []string{"CBOS"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "National ID"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},

		// Afrique Australe
		{
			Code: "ZA", Name: "Afrique du Sud", PhonePrefix: "+27", Region: RegionSouthernAfrica,
			Provider: "smileid", Regulations: []string{"SARB", "POPIA"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "SA ID", Pattern: `^[0-9]{13}$`},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "ZW", Name: "Zimbabwe", PhonePrefix: "+263", Region: RegionSouthernAfrica,
			Provider: "smileid", Regulations: []string{"RBZ"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "National ID"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "ZM", Name: "Zambie", PhonePrefix: "+260", Region: RegionSouthernAfrica,
			Provider: "smileid", Regulations: []string{"BOZ"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "NRC"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "BW", Name: "Botswana", PhonePrefix: "+267", Region: RegionSouthernAfrica,
			Provider: "smileid", Regulations: []string{"BOB"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "Omang"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "NA", Name: "Namibie", PhonePrefix: "+264", Region: RegionSouthernAfrica,
			Provider: "smileid", Regulations: []string{"BON"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "National ID"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "MW", Name: "Malawi", PhonePrefix: "+265", Region: RegionSouthernAfrica,
			Provider: "smileid", Regulations: []string{"RBM"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "National ID"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "MZ", Name: "Mozambique", PhonePrefix: "+258", Region: RegionSouthernAfrica,
			Provider: "smileid", Regulations: []string{"BOM"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "BI"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "LS", Name: "Lesotho", PhonePrefix: "+266", Region: RegionSouthernAfrica,
			Provider: "smileid", Regulations: []string{"CBL"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "National ID"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "SZ", Name: "Eswatini", PhonePrefix: "+268", Region: RegionSouthernAfrica,
			Provider: "smileid", Regulations: []string{"CBS"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "National ID"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "AO", Name: "Angola", PhonePrefix: "+244", Region: RegionSouthernAfrica,
			Provider: "smileid", Regulations: []string{"BNA"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "BI"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},

		// Afrique Centrale
		{
			Code: "CM", Name: "Cameroun", PhonePrefix: "+237", Region: RegionCentralAfrica,
			Provider: "smileid", Regulations: []string{"BEAC", "CEMAC"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNI"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "CD", Name: "République Démocratique du Congo", PhonePrefix: "+243", Region: RegionCentralAfrica,
			Provider: "smileid", Regulations: []string{"BEAC", "BCC"},
			DocTypes: []DocType{
				{Code: "VOTER_CARD", Name: "Carte d'électeur"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "CG", Name: "Congo-Brazzaville", PhonePrefix: "+242", Region: RegionCentralAfrica,
			Provider: "smileid", Regulations: []string{"BEAC", "CEMAC"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNI"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "CF", Name: "République Centrafricaine", PhonePrefix: "+236", Region: RegionCentralAfrica,
			Provider: "smileid", Regulations: []string{"BEAC", "CEMAC"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNI"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "TD", Name: "Tchad", PhonePrefix: "+235", Region: RegionCentralAfrica,
			Provider: "smileid", Regulations: []string{"BEAC", "CEMAC"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNI"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "GA", Name: "Gabon", PhonePrefix: "+241", Region: RegionCentralAfrica,
			Provider: "smileid", Regulations: []string{"BEAC", "CEMAC"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNI"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "GQ", Name: "Guinée Équatoriale", PhonePrefix: "+240", Region: RegionCentralAfrica,
			Provider: "smileid", Regulations: []string{"BEAC", "CEMAC"},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNI"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
		{
			Code: "ST", Name: "São Tomé-et-Príncipe", PhonePrefix: "+239", Region: RegionCentralAfrica,
			Provider: "smileid", Regulations: []string{},
			DocTypes: []DocType{
				{Code: "NATIONAL_ID", Name: "CNI"},
				{Code: "PASSPORT", Name: "Passeport"},
			},
		},
	}

	m := make(map[string]Country, len(countries))
	for _, c := range countries {
		m[strings.ToUpper(c.Code)] = c
	}
	return m
}

func GetCountry(code string) (*Country, bool) {
	c, ok := registry[strings.ToUpper(code)]
	if !ok {
		return nil, false
	}
	return &c, true
}

func GetProvider(countryCode string) string {
	c, ok := GetCountry(countryCode)
	if !ok {
		return "sumsub"
	}
	return c.Provider
}

func IsDocTypeValid(countryCode, docType string) bool {
	c, ok := GetCountry(countryCode)
	if !ok {
		return false
	}
	docType = strings.ToUpper(docType)
	for _, dt := range c.DocTypes {
		if strings.ToUpper(dt.Code) == docType {
			return true
		}
	}
	return false
}

func ValidateDocNumber(countryCode, docType, number string) bool {
	if number == "" {
		return true
	}

	c, ok := GetCountry(countryCode)
	if !ok {
		return false
	}

	docType = strings.ToUpper(docType)
	for _, dt := range c.DocTypes {
		if strings.ToUpper(dt.Code) != docType {
			continue
		}
		if dt.Pattern == "" {
			return true
		}
		matched, err := regexp.MatchString(dt.Pattern, number)
		return err == nil && matched
	}
	return false
}

func GetPhonePrefix(countryCode string) string {
	c, ok := GetCountry(countryCode)
	if !ok {
		return ""
	}
	return c.PhonePrefix
}

func ListCountries() []Country {
	result := make([]Country, 0, len(registry))
	for _, c := range registry {
		result = append(result, c)
	}
	return result
}

func ListByRegion(region string) []Country {
	region = strings.ToUpper(region)
	result := make([]Country, 0)
	for _, c := range registry {
		if c.Region == region {
			result = append(result, c)
		}
	}
	return result
}

// AllAfricanCountryCodes returns ISO 3166-1 alpha-2 codes for all 55 AU member states.
func AllAfricanCountryCodes() []string {
	return []string{
		"DZ", "AO", "BJ", "BW", "BF", "BI", "CV", "CM", "CF", "TD",
		"KM", "CG", "CD", "CI", "DJ", "EG", "GQ", "ER", "SZ", "ET",
		"GA", "GM", "GH", "GN", "GW", "KE", "LS", "LR", "LY", "MG",
		"MW", "ML", "MR", "MU", "MA", "MZ", "NA", "NE", "NG", "RW",
		"ST", "SN", "SC", "SL", "SO", "ZA", "SS", "SD", "TZ", "TG",
		"TN", "UG", "ZM", "ZW", "EH",
	}
}
