package notification

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

type SMSChannel interface {
	Channel
	IsAvailable() bool
}

type ConsoleSMS struct {
	logger *zap.Logger
}

func NewConsoleSMS(logger *zap.Logger) *ConsoleSMS {
	return &ConsoleSMS{logger: logger}
}

func (s *ConsoleSMS) Name() string { return "console_sms" }

func (s *ConsoleSMS) IsAvailable() bool { return true }

func (s *ConsoleSMS) Send(ctx context.Context, notif Notification) error {
	msg := s.formatMessage(notif)
	s.logger.Info("📱 SMS envoyé",
		zap.String("to", notif.Phone),
		zap.String("message", msg),
	)
	return nil
}

func (s *ConsoleSMS) formatMessage(notif Notification) string {
	switch notif.Event {
	case EventKYCInitiated:
		return fmt.Sprintf(
			"KYC %s initiée. Votre identifiant est %s. Statut: %s.",
			notif.CountryCode, notif.VerificationID, notif.Status,
		)
	case EventKYCApproved:
		return fmt.Sprintf(
			"✅ Votre vérification KYC %s est approuvée (score: %.0f%%).",
			notif.VerificationID, notif.Score*100,
		)
	case EventKYCRejected:
		return fmt.Sprintf(
			"❌ Votre vérification KYC %s a échoué. Contactez le support.",
			notif.VerificationID,
		)
	case EventKYCPendingReview:
		return fmt.Sprintf(
			"⏳ Votre vérification KYC %s est en examen manuel.",
			notif.VerificationID,
		)
	default:
		return fmt.Sprintf(
			"KYC %s mise à jour: %s", notif.VerificationID, notif.Status,
		)
	}
}
