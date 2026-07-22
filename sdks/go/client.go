package datakeys

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const (
	sandboxURL    = "http://localhost:8081"
	productionURL = "https://api.datakeys.africa"
	sdkVersion    = "1.0.0"
)

var codeRegexp = regexp.MustCompile(`KYC_[A-Z0-9_]+`)

type Config struct {
	BaseURL    string
	Timeout    time.Duration
	MaxRetries int
}

type Client struct {
	apiKey     string
	baseURL    string
	Livemode   bool
	httpClient *http.Client
	maxRetries int
}

func newClient(apiKey string, cfg Config) (*Client, error) {
	if strings.TrimSpace(apiKey) == "" {
		return nil, &KYCError{
			Code:    ErrAuthMissing,
			Message: "API key manquante",
			Status:  401,
		}
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	maxRetries := cfg.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	livemode := strings.HasPrefix(apiKey, "dk_live_")
	baseURL := cfg.BaseURL
	if baseURL == "" {
		if livemode {
			baseURL = productionURL
		} else {
			baseURL = sandboxURL
		}
	}

	return &Client{
		apiKey:     apiKey,
		baseURL:    baseURL,
		Livemode:   livemode,
		maxRetries: maxRetries,
		httpClient: &http.Client{Timeout: timeout},
	}, nil
}

func (c *Client) do(ctx context.Context, method, path string, body any, idempotencyKey string) ([]byte, error) {
	var bodyBytes []byte
	if body != nil {
		var err error
		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
	}

	var lastErr error

	for attempt := 0; attempt < c.maxRetries; attempt++ {
		req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", c.apiKey)
		req.Header.Set("X-SDK-Version", sdkVersion)
		req.Header.Set("X-SDK-Lang", "go")
		if idempotencyKey != "" {
			req.Header.Set("Idempotency-Key", idempotencyKey)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = &KYCError{Code: ErrNetwork, Message: err.Error()}
			if attempt < c.maxRetries-1 {
				c.backoff(attempt)
			}
			continue
		}

		data, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode >= 500 {
			lastErr = &KYCError{Code: ErrServerError, Message: "Erreur serveur", Status: resp.StatusCode}
			if attempt < c.maxRetries-1 {
				c.backoff(attempt)
			}
			continue
		}

		if resp.StatusCode >= 400 {
			var errResp struct {
				Error *string `json:"error"`
			}
			json.Unmarshal(data, &errResp)
			msg := "Erreur API"
			if errResp.Error != nil {
				msg = *errResp.Error
			}
			return nil, &KYCError{
				Code:    extractCode(msg),
				Message: msg,
				Status:  resp.StatusCode,
			}
		}

		return data, nil
	}

	return nil, lastErr
}

func (c *Client) backoff(attempt int) {
	base := 500 * time.Millisecond
	delay := time.Duration(float64(base) * math.Pow(2, float64(attempt)))
	jitter := time.Duration(rand.Float64() * float64(delay) * 0.3)
	if delay > 10*time.Second {
		delay = 10 * time.Second
	}
	time.Sleep(delay + jitter)
}

func extractCode(s string) ErrorCode {
	m := codeRegexp.FindString(s)
	if m == "" {
		return ErrUnknown
	}
	return ErrorCode(m)
}
