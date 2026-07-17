package resilience

import (
	"time"

	"github.com/sony/gobreaker"
	"go.uber.org/zap"
)

type CBConfig struct {
	Name        string
	MaxFailures uint32
	Timeout     time.Duration
	MaxRequests uint32
}

func DefaultCBConfig(name string) CBConfig {
	return CBConfig{
		Name:        name,
		MaxFailures: 5,
		Timeout:     30 * time.Second,
		MaxRequests: 2,
	}
}

type CircuitBreaker struct {
	cb     *gobreaker.CircuitBreaker
	logger *zap.Logger
}

func NewCircuitBreaker(cfg CBConfig, logger *zap.Logger) *CircuitBreaker {
	settings := gobreaker.Settings{
		Name:        cfg.Name,
		MaxRequests: cfg.MaxRequests,
		Interval:    60 * time.Second,
		Timeout:     cfg.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= cfg.MaxFailures
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			logger.Warn("circuit breaker state change",
				zap.String("provider", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()),
			)
		},
	}
	return &CircuitBreaker{
		cb:     gobreaker.NewCircuitBreaker(settings),
		logger: logger,
	}
}

func (cb *CircuitBreaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	return cb.cb.Execute(fn)
}

func (cb *CircuitBreaker) IsOpen() bool {
	return cb.cb.State() == gobreaker.StateOpen
}

func (cb *CircuitBreaker) State() string {
	return cb.cb.State().String()
}
