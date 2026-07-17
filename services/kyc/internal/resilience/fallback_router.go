package resilience

import (
	"context"
	"errors"
	"fmt"

	"github.com/datakeys/kyc-service/internal"
	"go.uber.org/zap"
)

var ErrAllProvidersDown = errors.New("tous les providers KYC sont down")

type FallbackRouter struct {
	providers []*ResilientProvider
	logger    *zap.Logger
}

func NewFallbackRouter(providers []*ResilientProvider, logger *zap.Logger) *FallbackRouter {
	return &FallbackRouter{providers: providers, logger: logger}
}

func (fr *FallbackRouter) Verify(ctx context.Context, req internal.ProviderRequest) (*internal.ProviderResult, string, error) {
	var lastErr error
	for _, p := range fr.providers {
		if !p.IsAvailable() {
			fr.logger.Warn("provider circuit open, skipping",
				zap.String("provider", p.Name()))
			continue
		}

		result, err := p.Verify(ctx, req)
		if err == nil {
			return result, p.Name(), nil
		}

		fr.logger.Warn("provider failed, trying next",
			zap.String("provider", p.Name()),
			zap.Error(err))
		lastErr = err
	}

	return nil, "", fmt.Errorf("tous les providers KYC sont indisponibles: %w", lastErr)
}

func (fr *FallbackRouter) ProviderStatuses() map[string]string {
	statuses := make(map[string]string)
	for _, p := range fr.providers {
		statuses[p.Name()] = p.cb.State()
	}
	return statuses
}
