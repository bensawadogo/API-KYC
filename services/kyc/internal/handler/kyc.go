package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/datakeys/kyc-service/internal"
	"github.com/datakeys/kyc-service/internal/countries"
	"github.com/datakeys/kyc-service/internal/model"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type KYCHandler struct {
	service   internal.KYCServiceInterface
	validate  *validator.Validate
	logger    *zap.Logger
	smileKey  string
	sumsubKey string
	redis     *redis.Client
}

func NewKYCHandler(
	service internal.KYCServiceInterface,
	logger *zap.Logger,
	smileIDKey, sumsubKey string,
	rdb *redis.Client,
) *KYCHandler {
	return &KYCHandler{
		service:   service,
		validate:  validator.New(),
		logger:    logger,
		smileKey:  smileIDKey,
		sumsubKey: sumsubKey,
		redis:     rdb,
	}
}

func (h *KYCHandler) Initiate(c fiber.Ctx) error {
	var req model.InitiateKYCRequest
	if err := c.Bind().Body(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("invalid request body"))
	}

	if err := h.validate.Struct(req); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(ErrorResponse(err.Error()))
	}

	resp, err := h.service.Initiate(c.Context(), &req)
	if err != nil {
		h.logger.Error("initiate kyc failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.Status(fiber.StatusCreated).JSON(SuccessResponse(resp))
}

func (h *KYCHandler) GetStatus(c fiber.Ctx) error {
	verificationID := c.Params("verification_id")
	if _, err := uuid.Parse(verificationID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("invalid verification_id"))
	}

	result, err := h.service.GetStatus(c.Context(), verificationID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(ErrorResponse("verification not found"))
		}
		h.logger.Error("get status failed", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(ErrorResponse(err.Error()))
	}

	return c.JSON(SuccessResponse(result))
}

func (h *KYCHandler) ListCountries(c fiber.Ctx) error {
	all := countries.ListCountries()
	type countryDTO struct {
		Code        string              `json:"code"`
		Name        string              `json:"name"`
		PhonePrefix string              `json:"phone_prefix"`
		Region      string              `json:"region"`
		Provider    string              `json:"provider"`
		DocTypes    []countries.DocType `json:"doc_types"`
		Regulations []string            `json:"regulations"`
	}

	dtos := make([]countryDTO, 0, len(all))
	for _, country := range all {
		dtos = append(dtos, countryDTO{
			Code:        country.Code,
			Name:        country.Name,
			PhonePrefix: country.PhonePrefix,
			Region:      country.Region,
			Provider:    country.Provider,
			DocTypes:    country.DocTypes,
			Regulations: country.Regulations,
		})
	}

	return c.JSON(SuccessResponse(dtos))
}

func (h *KYCHandler) ListDocTypes(c fiber.Ctx) error {
	code := strings.ToUpper(c.Params("code"))
	country, ok := countries.GetCountry(code)
	if !ok {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse("country not supported"))
	}

	return c.JSON(SuccessResponse(country.DocTypes))
}

func (h *KYCHandler) SmileIDWebhook(c fiber.Ctx) error {
	body := c.Body()
	if !h.verifySmileIDSignature(body, c.Get("X-Smile-Signature")) {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("invalid signature"))
	}

	verificationID := extractVerificationID(body)
	if verificationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("verification_id missing"))
	}

	if h.redis != nil {
		eventID := c.Get("X-Idempotency-Key")
		if eventID == "" {
			eventID = c.Get("X-Smile-Event-Id")
		}
		if eventID != "" {
			dedupKey := fmt.Sprintf("webhook:dedup:smileid:%s", eventID)
			set, err := h.redis.SetNX(c.Context(), dedupKey, "1", 72*time.Hour).Result()
			if err == nil && !set {
				h.logger.Info("smileid webhook deduplicated",
					zap.String("event_id", eventID))
				return c.JSON(SuccessResponse(map[string]string{"status": "already_processed"}))
			}
		}
	}

	go func(id string) {
		ctx := context.Background()
		if err := h.service.Process(ctx, id); err != nil {
			h.logger.Error("smileid webhook process failed",
				zap.String("verification_id", id),
				zap.Error(err),
			)
		}
	}(verificationID)

	return c.JSON(SuccessResponse(map[string]string{"status": "accepted"}))
}

func (h *KYCHandler) SumSubWebhook(c fiber.Ctx) error {
	body := c.Body()
	if !h.verifySumSubSignature(body, c.Get("X-Payload-Digest")) {
		return c.Status(fiber.StatusUnauthorized).JSON(ErrorResponse("invalid signature"))
	}

	verificationID := extractVerificationID(body)
	if verificationID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(ErrorResponse("verification_id missing"))
	}

	if h.redis != nil {
		eventID := c.Get("X-Idempotency-Key")
		if eventID == "" {
			eventID = c.Get("X-Sumsub-Event-Id")
		}
		if eventID != "" {
			dedupKey := fmt.Sprintf("webhook:dedup:sumsub:%s", eventID)
			set, err := h.redis.SetNX(c.Context(), dedupKey, "1", 72*time.Hour).Result()
			if err == nil && !set {
				h.logger.Info("sumsub webhook deduplicated",
					zap.String("event_id", eventID))
				return c.JSON(SuccessResponse(map[string]string{"status": "already_processed"}))
			}
		}
	}

	go func(id string) {
		ctx := context.Background()
		if err := h.service.Process(ctx, id); err != nil {
			h.logger.Error("sumsub webhook process failed",
				zap.String("verification_id", id),
				zap.Error(err),
			)
		}
	}(verificationID)

	return c.JSON(SuccessResponse(map[string]string{"status": "accepted"}))
}

func (h *KYCHandler) verifySmileIDSignature(body []byte, signature string) bool {
	if h.smileKey == "" || signature == "" {
		return true
	}
	mac := hmac.New(sha256.New, []byte(h.smileKey))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

func (h *KYCHandler) verifySumSubSignature(body []byte, digest string) bool {
	if h.sumsubKey == "" || digest == "" {
		return true
	}
	mac := hmac.New(sha256.New, []byte(h.sumsubKey))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(digest))
}

func extractVerificationID(body []byte) string {
	var payload map[string]interface{}
	if err := json.Unmarshal(body, &payload); err != nil {
		return ""
	}

	if id, ok := payload["user_id"].(string); ok && id != "" {
		return id
	}
	if id, ok := payload["externalUserId"].(string); ok && id != "" {
		return id
	}
	if data, ok := payload["data"].(map[string]interface{}); ok {
		if id, ok := data["user_id"].(string); ok {
			return id
		}
		if id, ok := data["externalUserId"].(string); ok {
			return id
		}
	}
	return ""
}
