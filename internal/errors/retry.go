package errors

import (
	"context"
	"math"
	"time"
)

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxAttempts int
	InitialWait time.Duration
	MaxWait     time.Duration
	Multiplier  float64
}

// DefaultRetryConfig returns sensible defaults
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts: 3,
		InitialWait: 1 * time.Second,
		MaxWait:     30 * time.Second,
		Multiplier:  2.0,
	}
}

// Retry executes fn with exponential backoff
func Retry(ctx context.Context, config RetryConfig, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry validation errors
		if IsType(err, ErrorTypeValidation) {
			return err
		}

		// Don't wait after last attempt
		if attempt == config.MaxAttempts-1 {
			break
		}

		// Calculate wait time with exponential backoff
		waitTime := time.Duration(float64(config.InitialWait) * math.Pow(config.Multiplier, float64(attempt)))
		if waitTime > config.MaxWait {
			waitTime = config.MaxWait
		}

		// Wait with context cancellation support
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitTime):
		}
	}

	return lastErr
}
