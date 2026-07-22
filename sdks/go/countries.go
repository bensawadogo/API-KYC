package datakeys

import (
	"context"
	"encoding/json"
	"fmt"
)

type CountryDocType struct {
	Code    string  `json:"code"`
	Name    string  `json:"name"`
	Pattern *string `json:"pattern,omitempty"`
}

type Country struct {
	Code        string          `json:"code"`
	Name        string          `json:"name"`
	PhonePrefix string          `json:"phone_prefix"`
	Region      string          `json:"region"`
	DocTypes    []CountryDocType `json:"doc_types"`
	Provider    string          `json:"provider"`
}

type CountriesService struct {
	client *Client
}

func (s *CountriesService) List(ctx context.Context) ([]Country, error) {
	data, err := s.client.do(ctx, "GET", "/v1/kyc/countries", nil, "")
	if err != nil {
		return nil, err
	}

	var res apiResponse[[]Country]
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("parse countries: %w", err)
	}
	if res.Data == nil {
		return nil, fmt.Errorf("empty countries response")
	}
	return *res.Data, nil
}

func (s *CountriesService) DocTypes(ctx context.Context, countryCode string) ([]CountryDocType, error) {
	if countryCode == "" {
		return nil, fmt.Errorf("countryCode requis")
	}
	data, err := s.client.do(ctx, "GET", "/v1/kyc/countries/"+countryCode+"/doctypes", nil, "")
	if err != nil {
		return nil, err
	}

	var res apiResponse[[]CountryDocType]
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("parse doctypes: %w", err)
	}
	if res.Data == nil {
		return nil, fmt.Errorf("empty doctypes response")
	}
	return *res.Data, nil
}
