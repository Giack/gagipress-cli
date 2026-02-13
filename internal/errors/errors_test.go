package errors

import (
	"errors"
	"testing"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *AppError
		contains string
	}{
		{
			name: "simple error",
			err: &AppError{
				Type:    ErrorTypeValidation,
				Message: "invalid input",
			},
			contains: "validation: invalid input",
		},
		{
			name: "wrapped error",
			err: &AppError{
				Type:    ErrorTypeAPI,
				Message: "API call failed",
				Err:     errors.New("connection refused"),
			},
			contains: "api: API call failed: connection refused",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.err.Error()
			if errMsg != tt.contains {
				t.Errorf("Error() = %q, want %q", errMsg, tt.contains)
			}
		})
	}
}

func TestNew(t *testing.T) {
	err := New(ErrorTypeValidation, "invalid input")

	if err.Type != ErrorTypeValidation {
		t.Errorf("expected type %s, got %s", ErrorTypeValidation, err.Type)
	}

	if err.Message != "invalid input" {
		t.Errorf("expected message 'invalid input', got %s", err.Message)
	}

	if err.Err != nil {
		t.Error("expected no wrapped error")
	}
}

func TestWrap(t *testing.T) {
	original := errors.New("original error")
	wrapped := Wrap(original, ErrorTypeAPI, "API call failed")

	if wrapped.Err != original {
		t.Error("Wrap did not preserve original error")
	}

	if wrapped.Type != ErrorTypeAPI {
		t.Errorf("expected type %s, got %s", ErrorTypeAPI, wrapped.Type)
	}

	if !errors.Is(wrapped, original) {
		t.Error("Wrapped error should unwrap to original")
	}
}

func TestIsType(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		errType  ErrorType
		expected bool
	}{
		{
			name:     "matching type",
			err:      New(ErrorTypeValidation, "test"),
			errType:  ErrorTypeValidation,
			expected: true,
		},
		{
			name:     "non-matching type",
			err:      New(ErrorTypeValidation, "test"),
			errType:  ErrorTypeAPI,
			expected: false,
		},
		{
			name:     "non-AppError",
			err:      errors.New("regular error"),
			errType:  ErrorTypeValidation,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsType(tt.err, tt.errType)
			if result != tt.expected {
				t.Errorf("IsType() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	original := errors.New("original")
	wrapped := Wrap(original, ErrorTypeAPI, "wrapped")

	unwrapped := wrapped.Unwrap()
	if unwrapped != original {
		t.Error("Unwrap did not return original error")
	}
}
