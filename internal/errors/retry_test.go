package errors

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetry_Success(t *testing.T) {
	attempts := 0
	fn := func() error {
		attempts++
		if attempts < 2 {
			return errors.New("temporary error")
		}
		return nil
	}

	config := RetryConfig{
		MaxAttempts: 3,
		InitialWait: 10 * time.Millisecond,
		MaxWait:     100 * time.Millisecond,
		Multiplier:  2.0,
	}

	err := Retry(context.Background(), config, fn)

	if err != nil {
		t.Errorf("Retry() failed: %v", err)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetry_AllAttemptsFail(t *testing.T) {
	attempts := 0
	fn := func() error {
		attempts++
		return errors.New("persistent error")
	}

	config := RetryConfig{
		MaxAttempts: 3,
		InitialWait: 10 * time.Millisecond,
		MaxWait:     100 * time.Millisecond,
		Multiplier:  2.0,
	}

	err := Retry(context.Background(), config, fn)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetry_ValidationError_NoRetry(t *testing.T) {
	attempts := 0
	fn := func() error {
		attempts++
		return New(ErrorTypeValidation, "invalid input")
	}

	config := RetryConfig{
		MaxAttempts: 3,
		InitialWait: 10 * time.Millisecond,
		MaxWait:     100 * time.Millisecond,
		Multiplier:  2.0,
	}

	err := Retry(context.Background(), config, fn)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Should not retry validation errors
	if attempts != 1 {
		t.Errorf("Expected 1 attempt (no retry), got %d", attempts)
	}
}

func TestRetry_ContextCancellation(t *testing.T) {
	attempts := 0
	fn := func() error {
		attempts++
		return errors.New("error")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	config := RetryConfig{
		MaxAttempts: 3,
		InitialWait: 10 * time.Millisecond,
		MaxWait:     100 * time.Millisecond,
		Multiplier:  2.0,
	}

	err := Retry(ctx, config, fn)

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", err)
	}

	// Should execute at least once before checking context
	if attempts < 1 {
		t.Errorf("Expected at least 1 attempt, got %d", attempts)
	}
}

func TestRetry_ExponentialBackoff(t *testing.T) {
	config := RetryConfig{
		MaxAttempts: 3,
		InitialWait: 10 * time.Millisecond,
		MaxWait:     100 * time.Millisecond,
		Multiplier:  2.0,
	}

	attempts := 0
	startTime := time.Now()

	fn := func() error {
		attempts++
		return errors.New("error")
	}

	Retry(context.Background(), config, fn)

	duration := time.Since(startTime)

	// Should take at least: 10ms (first wait) + 20ms (second wait) = 30ms
	// Adding some buffer for execution time
	if duration < 25*time.Millisecond {
		t.Errorf("Expected at least 25ms duration, got %v", duration)
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxAttempts != 3 {
		t.Errorf("Expected MaxAttempts=3, got %d", config.MaxAttempts)
	}

	if config.InitialWait != 1*time.Second {
		t.Errorf("Expected InitialWait=1s, got %v", config.InitialWait)
	}

	if config.MaxWait != 30*time.Second {
		t.Errorf("Expected MaxWait=30s, got %v", config.MaxWait)
	}

	if config.Multiplier != 2.0 {
		t.Errorf("Expected Multiplier=2.0, got %v", config.Multiplier)
	}
}
