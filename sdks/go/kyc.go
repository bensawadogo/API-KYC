package datakeys

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type KYCService struct {
	client *Client
}

func (s *KYCService) Initiate(ctx context.Context, params InitiateParams, iKeys ...string) (*KYCVerification, error) {
	iKey := uuid.New().String()
	if len(iKeys) > 0 && iKeys[0] != "" {
		iKey = iKeys[0]
	}

	data, err := s.client.do(ctx, "POST", "/v1/kyc/initiate", params, iKey)
	if err != nil {
		return nil, err
	}

	return parseVerification(data)
}

func (s *KYCService) Retrieve(ctx context.Context, verificationID string) (*KYCVerification, error) {
	if verificationID == "" {
		return nil, fmt.Errorf("verificationID requis")
	}
	data, err := s.client.do(ctx, "GET", "/v1/kyc/status/"+verificationID, nil, "")
	if err != nil {
		return nil, err
	}
	return parseVerification(data)
}

func (s *KYCService) WaitForCompletion(ctx context.Context, verificationID string, maxWait time.Duration, opts ...WaitOption) (*KYCVerification, error) {
	if maxWait == 0 {
		maxWait = 2 * time.Minute
	}

	cfg := waitConfig{interval: 3 * time.Second}
	for _, o := range opts {
		o(&cfg)
	}

	deadline := time.Now().Add(maxWait)

	for time.Now().Before(deadline) {
		v, err := s.Retrieve(ctx, verificationID)
		if err != nil {
			return nil, err
		}

		if cfg.onPoll != nil {
			cfg.onPoll(v)
		}

		if v.IsTerminal() {
			return v, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(cfg.interval):
		}
	}

	return nil, fmt.Errorf("timeout: vérification %s toujours en attente après %s", verificationID, maxWait)
}

type waitConfig struct {
	interval time.Duration
	onPoll   func(*KYCVerification)
}

type WaitOption func(*waitConfig)

func WithInterval(d time.Duration) WaitOption {
	return func(c *waitConfig) { c.interval = d }
}

func WithPollCallback(fn func(*KYCVerification)) WaitOption {
	return func(c *waitConfig) { c.onPoll = fn }
}

func parseVerification(data []byte) (*KYCVerification, error) {
	var res apiResponse[KYCVerification]
	if err := json.Unmarshal(data, &res); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}
	if res.Data == nil {
		msg := "erreur inconnue"
		if res.Error != nil {
			msg = *res.Error
		}
		return nil, fmt.Errorf("%s", msg)
	}
	return res.Data, nil
}
