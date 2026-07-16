package internal

import (
	"errors"
	"fmt"
	"testing"
)

func TestIsNotFoundError(t *testing.T) {
	t.Run("returns true for NotFoundError", func(t *testing.T) {
		err := NewNotFoundError(1, "Product")
		if !IsNotFoundError(err) {
			t.Errorf("IsNotFoundError(err) = false; want true")
		}
	})

	t.Run("returns true for wrapped NotFoundError", func(t *testing.T) {
		err := fmt.Errorf("wrapped: %w", NewNotFoundError(1, "Product"))
		if !IsNotFoundError(err) {
			t.Errorf("IsNotFoundError(err) = false; want true")
		}
	})

	t.Run("returns false for other errors", func(t *testing.T) {
		err := errors.New("some other error")
		if IsNotFoundError(err) {
			t.Errorf("IsNotFoundError(err) = true; want false")
		}
	})

	t.Run("returns false for nil error", func(t *testing.T) {
		if IsNotFoundError(nil) {
			t.Errorf("IsNotFoundError(nil) = true; want false")
		}
	})
}

func TestIsValidationError(t *testing.T) {
	t.Run("returns true for ValidationError", func(t *testing.T) {
		err := NewValidationErr("invalid input")
		if !IsValidationError(err) {
			t.Errorf("IsValidationError(err) = false; want true")
		}
	})

	t.Run("returns true for wrapped ValidationError", func(t *testing.T) {
		err := fmt.Errorf("wrapped: %w", NewValidationErr("invalid input"))
		if !IsValidationError(err) {
			t.Errorf("IsValidationError(err) = false; want true")
		}
	})

	t.Run("returns false for other errors", func(t *testing.T) {
		err := errors.New("some other error")
		if IsValidationError(err) {
			t.Errorf("IsValidationError(err) = true; want false")
		}
	})

	t.Run("returns false for nil error", func(t *testing.T) {
		if IsValidationError(nil) {
			t.Errorf("IsValidationError(nil) = true; want false")
		}
	})
}

func TestNotFoundError(t *testing.T) {
	err := NewNotFoundError(123, "Product")

	t.Run("Error message", func(t *testing.T) {
		expected := "Product not found with ID: 123"
		if err.Error() != expected {
			t.Errorf("expected %q, got %q", expected, err.Error())
		}
	})

}
