package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/datakeys/kyc-service/config"
	"github.com/datakeys/kyc-service/internal"
	"github.com/datakeys/kyc-service/internal/countries"
	"github.com/datakeys/kyc-service/internal/model"
	"github.com/datakeys/kyc-service/internal/observability"
	"github.com/datakeys/kyc-service/internal/resilience"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type KYCService struct {
	repo      internal.KYCRepository
	storage   internal.DocumentStorage
	providers map[string]internal.IdentityProvider
	router    *resilience.FallbackRouter
	webhook   internal.WebhookSender
	registry  internal.CountryRegistry
	cfg       config.Config
	logger    *zap.Logger
	aml       internal.AMLChecker
	amlRepo   internal.AMLRepository
	dlq       resilience.DLQInterface
}

func NewKYCService(
	repo internal.KYCRepository,
	storage internal.DocumentStorage,
	providers map[string]internal.IdentityProvider,
	webhook internal.WebhookSender,
	registry internal.CountryRegistry,
	cfg config.Config,
	logger *zap.Logger,
	aml internal.AMLChecker,
	amlRepo internal.AMLRepository,
	router *resilience.FallbackRouter,
	dlq resilience.DLQInterface,
) *KYCService {
	return &KYCService{
		repo:      repo,
		storage:   storage,
		providers: providers,
		router:    router,
		webhook:   webhook,
		registry:  registry,
		cfg:       cfg,
		logger:    logger,
		aml:       aml,
		amlRepo:   amlRepo,
		dlq:       dlq,
	}
}

func (s *KYCService) Initiate(ctx context.Context, req *model.InitiateKYCRequest) (*model.InitiateKYCResponse, error) {
	if s.cfg.Compliance.ConsentRequired && !req.Consent {
		return nil, fmt.Errorf("consent is required for KYC verification")
	}

	countryCode := strings.ToUpper(req.CountryCode)
	if _, ok := s.registry.GetCountry(countryCode); !ok {
		return nil, fmt.Errorf("unsupported country: %s", countryCode)
	}

	if !s.registry.IsDocTypeValid(countryCode, req.DocType) {
		return nil, fmt.Errorf("unsupported document type %s for country %s", req.DocType, countryCode)
	}

	if req.DocNumber != "" && !s.registry.ValidateDocNumber(countryCode, req.DocType, req.DocNumber) {
		return nil, fmt.Errorf("invalid document number format")
	}

	flags := make([]string, 0)
	expectedPrefix := countries.GetPhonePrefix(countryCode)
	if expectedPrefix != "" && !strings.HasPrefix(req.Phone, expectedPrefix) {
		flags = append(flags, model.FlagCountryMismatch)
	}

	exists, err := s.repo.ExistsApproved(ctx, req.Phone, countryCode)
	if err != nil {
		return nil, fmt.Errorf("check existing verification: %w", err)
	}
	if exists {
		flags = append(flags, model.FlagDuplicatePhone)
	}

	providerName := s.selectProvider(countryCode)
	verificationID := uuid.New().String()

	var amlResult *internal.AMLResult
	if s.aml != nil {
		var amlErr error
		amlResult, amlErr = s.aml.Check(ctx, internal.AMLRequest{
			FullName:    req.FullName,
			CountryCode: countryCode,
			DateOfBirth: "",
			DocNumber:   req.DocNumber,
		})
		if amlErr != nil {
			s.logger.Warn("aml screening failed, proceeding without",
				zap.String("verification_id", verificationID),
				zap.Error(amlErr),
			)
			flags = append(flags, model.FlagAMLCheckFailed)
		}
	} else {
		s.logger.Warn("aml checker not configured")
		flags = append(flags, model.FlagAMLCheckFailed)
	}

	if amlResult != nil {
		if amlResult.IsSanctioned {
			s.logger.Info("aml sanctions match, rejecting",
				zap.String("verification_id", verificationID),
				zap.Float64("score", amlResult.Score),
				zap.Int("matches", len(amlResult.Matches)),
			)
			result := &model.VerificationResult{
				VerificationID: verificationID,
				Phone:          req.Phone,
				CountryCode:    countryCode,
				DocType:        strings.ToUpper(req.DocType),
				DocNumber:      req.DocNumber,
				Status:         model.StatusRejected,
				Provider:       providerName,
				Flags:          append(flags, model.FlagSanctionsMatch),
				CallbackURL:    req.CallbackURL,
				Consent:        req.Consent,
				IsSanctioned:   true,
				AMLScore:       amlResult.Score,
			}
			if saveErr := s.repo.CreateVerification(ctx, result); saveErr != nil {
				return nil, fmt.Errorf("create verification after sanctions rejection: %w", saveErr)
			}
			if saveAMLErr := s.amlRepo.SaveAMLResult(ctx, verificationID, amlResult); saveAMLErr != nil {
				s.logger.Warn("failed to save aml result",
					zap.String("verification_id", verificationID),
					zap.Error(saveAMLErr),
				)
			}
			return &model.InitiateKYCResponse{
				VerificationID: verificationID,
				Status:         model.StatusRejected,
				Message:        "Verification cannot proceed due to sanctions screening result.",
			}, nil
		}

		if amlResult.IsPEP {
			flags = append(flags, model.FlagPEPDetected)
		}

		if saveAMLErr := s.amlRepo.SaveAMLResult(ctx, verificationID, amlResult); saveAMLErr != nil {
			s.logger.Warn("failed to save aml result",
				zap.String("verification_id", verificationID),
				zap.Error(saveAMLErr),
			)
		}
	}

	uploadURL, err := s.storage.GenerateUploadURL(ctx, verificationID, req.DocType)
	if err != nil {
		return nil, fmt.Errorf("generate upload url: %w", err)
	}

	expiresAt := time.Now().UTC().Add(time.Duration(s.cfg.KYC.SessionTTLSeconds) * time.Second)
	verification := &model.VerificationResult{
		VerificationID: verificationID,
		Phone:          req.Phone,
		CountryCode:    countryCode,
		DocType:        strings.ToUpper(req.DocType),
		DocNumber:      req.DocNumber,
		Status:         model.StatusPending,
		Provider:       providerName,
		Flags:          flags,
		CallbackURL:    req.CallbackURL,
		Consent:        req.Consent,
		ExpiresAt:      expiresAt.Format(time.RFC3339),
	}
	if amlResult != nil {
		verification.IsSanctioned = amlResult.IsSanctioned
		verification.IsPEP = amlResult.IsPEP
		verification.AMLScore = amlResult.Score
	}

	if err := s.repo.CreateVerification(ctx, verification); err != nil {
		return nil, fmt.Errorf("create verification: %w", err)
	}

	observability.KYCInitiated.WithLabelValues(countryCode, verification.DocType).Inc()

	if s.aml != nil {
		for _, flag := range flags {
			if flag == model.FlagSanctionsMatch || flag == model.FlagPEPDetected || flag == model.FlagAMLCheckFailed {
				observability.KYCAMLFlagged.WithLabelValues(flag, countryCode).Inc()
			}
		}
	}

	s.logger.Info("kyc verification initiated",
		zap.String("verification_id", verificationID),
		zap.String("country_code", countryCode),
		zap.String("doc_type", verification.DocType),
		zap.String("provider", providerName),
		zap.Strings("flags", flags),
	)

	return &model.InitiateKYCResponse{
		VerificationID: verificationID,
		Status:         model.StatusPending,
		UploadURL:      uploadURL,
		ExpiresIn:      s.cfg.KYC.SessionTTLSeconds,
		Provider:       providerName,
		Message:        "Verification session created. Upload document to proceed.",
	}, nil
}

func (s *KYCService) Process(ctx context.Context, verificationID string) error {
	processCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	ctx = processCtx

	start := time.Now()

	verification, err := s.repo.GetVerification(ctx, verificationID)
	if err != nil {
		return err
	}

	if verification.Status != model.StatusPending {
		return fmt.Errorf("verification %s is not pending (status=%s)", verificationID, verification.Status)
	}

	docData, err := s.storage.GetDocument(ctx, verificationID)
	if err != nil {
		return fmt.Errorf("get document: %w", err)
	}

	if err := s.repo.UpdateStatus(ctx, verificationID, model.StatusProcessing, 0, verification.Flags, verification.Provider); err != nil {
		return fmt.Errorf("set processing status: %w", err)
	}

	providerReq := internal.ProviderRequest{
		VerificationID: verificationID,
		CountryCode:    verification.CountryCode,
		DocType:        verification.DocType,
		DocNumber:      verification.DocNumber,
		DocData:        docData,
		Phone:          verification.Phone,
	}

	var result *internal.ProviderResult
	var usedProvider string

	if s.router != nil {
		result, usedProvider, err = s.router.Verify(ctx, providerReq)
		if err != nil {
			if s.dlq != nil {
				if dlqErr := s.dlq.Enqueue(ctx, verificationID); dlqErr != nil {
					s.logger.Error("failed to enqueue to DLQ",
						zap.String("verification_id", verificationID),
						zap.Error(dlqErr),
					)
				}
			}
			_ = s.repo.UpdateStatus(ctx, verificationID, model.StatusRejected, 0,
				[]string{model.FlagProviderUnavailable}, "none")
			return fmt.Errorf("provider unavailable: %w", err)
		}
		s.logger.Info("verification processed",
			zap.String("provider", usedProvider),
			zap.String("verification_id", verificationID))
	} else {
		provider, ok := s.providers[verification.Provider]
		if !ok {
			provider = s.providers["local"]
		}
		result, err = provider.Verify(ctx, providerReq)
		if err != nil {
			return fmt.Errorf("provider verify: %w", err)
		}
		usedProvider = result.Provider
	}

	flags := mergeFlags(verification.Flags, result.Flags)
	status := resolveStatus(result.Score, result.Approved, s.cfg.KYC.ScoreThreshold)

	if err := s.repo.UpdateStatus(ctx, verificationID, status, result.Score, flags, usedProvider); err != nil {
		return fmt.Errorf("update verification result: %w", err)
	}

	observability.KYCCompleted.WithLabelValues(status, usedProvider, verification.CountryCode).Inc()
	observability.KYCDuration.WithLabelValues(usedProvider, verification.CountryCode).Observe(time.Since(start).Seconds())

	if status == model.StatusApproved {
		if err := s.storage.DeleteDocument(ctx, verificationID); err != nil {
			s.logger.Warn("failed to delete document after approval",
				zap.String("verification_id", verificationID),
				zap.Error(err),
			)
		}
	}

	if verification.CallbackURL != "" {
		callbackURL := verification.CallbackURL
		go s.sendWebhook(callbackURL, verificationID)
	}

	if s.cfg.Compliance.AuditEnabled {
		s.logger.Info("kyc verification processed",
			zap.String("verification_id", verificationID),
			zap.String("country_code", verification.CountryCode),
			zap.String("doc_type", verification.DocType),
			zap.String("provider", usedProvider),
			zap.Float64("score", result.Score),
			zap.Strings("flags", flags),
			zap.Int64("duration_ms", time.Since(start).Milliseconds()),
		)
	}

	return nil
}

func (s *KYCService) GetStatus(ctx context.Context, verificationID string) (*model.VerificationResult, error) {
	return s.repo.GetVerification(ctx, verificationID)
}

func (s *KYCService) selectProvider(countryCode string) string {
	preferred := s.registry.GetProvider(countryCode)
	if _, ok := s.providers[preferred]; ok {
		return preferred
	}

	fallbackOrder := []string{"smileid", "youverify", "sumsub", "local"}
	if s.cfg.KYC.DefaultProvider != "" {
		fallbackOrder = append([]string{s.cfg.KYC.DefaultProvider}, fallbackOrder...)
	}

	seen := make(map[string]bool)
	for _, name := range fallbackOrder {
		if seen[name] {
			continue
		}
		seen[name] = true
		if _, ok := s.providers[name]; ok {
			return name
		}
	}

	return "local"
}

func (s *KYCService) sendWebhook(callbackURL, verificationID string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	verification, err := s.repo.GetVerification(ctx, verificationID)
	if err != nil {
		s.logger.Error("webhook fetch verification failed",
			zap.String("verification_id", verificationID),
			zap.Error(err),
		)
		return
	}

	payload := &model.WebhookPayload{
		Event: "kyc.verification.completed",
		Data:  *verification,
	}

	if err := s.webhook.Send(ctx, callbackURL, payload); err != nil {
		s.logger.Error("webhook delivery failed",
			zap.String("verification_id", verificationID),
			zap.Error(err),
		)
	}
}

func resolveStatus(score float64, approved bool, threshold float64) string {
	if approved && score >= threshold {
		return model.StatusApproved
	}
	if score < 0.4 {
		return model.StatusRejected
	}
	return model.StatusManualReview
}

func mergeFlags(existing, incoming []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(existing)+len(incoming))
	for _, f := range append(existing, incoming...) {
		if f == "" || seen[f] {
			continue
		}
		seen[f] = true
		result = append(result, f)
	}
	return result
}

// ErrVerificationNotFound indicates the verification ID does not exist.
var ErrVerificationNotFound = errors.New("verification not found")
