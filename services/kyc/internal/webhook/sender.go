package webhook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/datakeys/kyc-service/internal/model"
)

type HTTPSender struct {
	client *http.Client
}

func NewHTTPSender() *HTTPSender {
	return &HTTPSender{
		client: &http.Client{Timeout: 15 * time.Second},
	}
}

func (s *HTTPSender) Send(ctx context.Context, url string, payload *model.WebhookPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if payload.Signature != "" {
		req.Header.Set("X-KYC-Signature", payload.Signature)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook send failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}
	return nil
}
