package handler

import (
	"fmt"

	"github.com/datakeys/kyc-service/internal"
	"github.com/datakeys/kyc-service/internal/provider"
	"github.com/datakeys/kyc-service/internal/seed"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/zap"
)

type SandboxHandler struct {
	profiles []seed.TestProfile
	apiKey   string
	baseURL  string
	logger   *zap.Logger
}

func NewSandboxHandler(profiles []seed.TestProfile, baseURL, apiKey string, logger *zap.Logger) *SandboxHandler {
	return &SandboxHandler{
		profiles: profiles,
		apiKey:   apiKey,
		baseURL:  baseURL,
		logger:   logger,
	}
}

type profileResponse struct {
	Phone       string `json:"phone"`
	CountryCode string `json:"country_code"`
	DocType     string `json:"doc_type"`
	DocNumber   string `json:"doc_number"`
	FullName    string `json:"full_name"`
	Scenario    string `json:"scenario"`
	Description string `json:"description"`
	CurlExample string `json:"curl_example"`
}

func (h *SandboxHandler) Profiles(c fiber.Ctx) error {
	profiles := make([]profileResponse, 0, len(h.profiles))
	for _, p := range h.profiles {
		profiles = append(profiles, profileResponse{
			Phone:       p.Phone,
			CountryCode: p.CountryCode,
			DocType:     p.DocType,
			DocNumber:   p.DocNumber,
			FullName:    p.FullName,
			Scenario:    p.Scenario,
			Description: p.Description,
			CurlExample: seed.CurlExample(p, h.baseURL, h.apiKey),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data": fiber.Map{
			"base_url":        h.baseURL,
			"profiles":        profiles,
			"total_profiles":  len(profiles),
			"quota_remaining": 1000,
		},
	})
}

func (h *SandboxHandler) Reset(c fiber.Ctx) error {
	h.logger.Info("sandbox quota reset")
	return c.JSON(fiber.Map{
		"success": true,
		"message": "Quota réinitialisé",
	})
}

type simulateRequest struct {
	Scenario string `json:"scenario"`
}

func (h *SandboxHandler) Simulate(c fiber.Ctx) error {
	var req simulateRequest
	if err := c.Bind().JSON(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   "corps JSON invalide",
		})
	}

	sp := provider.NewSandboxProvider()
	dummyReq := struct{ DocNumber string }{}
	switch req.Scenario {
	case "approved":
		dummyReq.DocNumber = "XXXX0000"
	case "rejected":
		dummyReq.DocNumber = "XXXX1111"
	case "sanctions":
		dummyReq.DocNumber = "XXXX2222"
	case "pep":
		dummyReq.DocNumber = "XXXX3333"
	case "expired_doc":
		dummyReq.DocNumber = "XXXX4444"
	default:
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"success": false,
			"error":   fmt.Sprintf("scénario inconnu: %s", req.Scenario),
		})
	}

	result, err := sp.Verify(c.Context(), internal.ProviderRequest{
		VerificationID: "sandbox-simulate",
		CountryCode:    "BF",
		DocType:        "NATIONAL_ID",
		DocNumber:      dummyReq.DocNumber,
		Phone:          "+22670000001",
	})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"success": false,
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    result,
	})
}
