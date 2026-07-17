package middleware

import (
	"strings"

	"github.com/datakeys/kyc-service/internal"
	"github.com/gofiber/fiber/v3"
)

type AuthErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func RequireAPIKey(repo internal.APIKeyRepository, requiredScope string) fiber.Handler {
	return func(c fiber.Ctx) error {
		rawKey := extractKey(c)
		if rawKey == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(AuthErrorResponse{
				Success: false,
				Error:   "KYC_AUTH_001: API key manquante",
			})
		}

		apiKey, err := repo.ValidateKey(c.Context(), rawKey)
		if err != nil || apiKey == nil {
			return c.Status(fiber.StatusUnauthorized).JSON(AuthErrorResponse{
				Success: false,
				Error:   "KYC_AUTH_002: API key invalide ou expirée",
			})
		}

		if requiredScope != "" && !hasScope(apiKey.Scopes, requiredScope) {
			return c.Status(fiber.StatusForbidden).JSON(AuthErrorResponse{
				Success: false,
				Error:   "KYC_AUTH_003: Scope insuffisant",
			})
		}

		c.Locals("api_key_id", apiKey.ID)
		c.Locals("client_name", apiKey.ClientName)
		c.Locals("rate_limit", apiKey.RateLimit)

		return c.Next()
	}
}

func extractKey(c fiber.Ctx) string {
	auth := c.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return c.Get("X-API-Key")
}

func hasScope(scopes []string, required string) bool {
	for _, s := range scopes {
		if s == required || s == "kyc:admin" {
			return true
		}
	}
	return false
}
