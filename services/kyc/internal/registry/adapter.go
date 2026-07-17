package registry

import (
	"github.com/datakeys/kyc-service/internal/countries"
)

type Adapter struct{}

func New() *Adapter {
	return &Adapter{}
}

func (a *Adapter) GetCountry(code string) (*countries.Country, bool) {
	return countries.GetCountry(code)
}

func (a *Adapter) IsDocTypeValid(countryCode, docType string) bool {
	return countries.IsDocTypeValid(countryCode, docType)
}

func (a *Adapter) ValidateDocNumber(countryCode, docType, number string) bool {
	return countries.ValidateDocNumber(countryCode, docType, number)
}

func (a *Adapter) GetProvider(countryCode string) string {
	return countries.GetProvider(countryCode)
}
