// Package config_test tests the environment variable helpers.
//
// Notice the package name is config_test, not config.
// This is called a "black box test" — we test the package
// from the outside, exactly as a caller would use it.
// This ensures we only test the public API, not internal details.
package config_test

import (
	"testing"

	"github.com/bytepharoh/rideflow/pkg/config"
)

// TestGetString covers all the cases GetString must handle.
func TestGetString(t *testing.T) {
	// t.Run creates a named sub-test.
	// Each sub-test runs independently and is reported separately.
	// This is the standard Go way to group related test cases.

	t.Run("returns value when env var is set", func(t *testing.T) {
		t.Setenv("TEST_KEY", "hello")

		result := config.GetString("TEST_KEY", "default")

		if result != "hello" {
			t.Errorf("expected %q, got %q", "hello", result)
		}
	})

	t.Run("returns default when env var is not set", func(t *testing.T) {
		// We do not set TEST_MISSING — it should not exist
		result := config.GetString("TEST_MISSING", "default")

		if result != "default" {
			t.Errorf("expected %q, got %q", "default", result)
		}
	})

	t.Run("returns default when env var is empty string", func(t *testing.T) {
		t.Setenv("TEST_EMPTY", "")

		result := config.GetString("TEST_EMPTY", "default")

		if result != "default" {
			t.Errorf("expected %q, got %q", "default", result)
		}
	})
}

// TestGetInt covers the integer parsing logic which has real failure modes.
func TestGetInt(t *testing.T) {
	t.Run("returns value when env var is a valid integer", func(t *testing.T) {
		t.Setenv("TEST_PORT", "8080")

		result, err := config.GetInt("TEST_PORT", 3000)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result != 8080 {
			t.Errorf("expected %d, got %d", 8080, result)
		}
	})

	t.Run("returns default when env var is not set", func(t *testing.T) {
		result, err := config.GetInt("TEST_PORT_MISSING", 3000)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result != 3000 {
			t.Errorf("expected %d, got %d", 3000, result)
		}
	})

	t.Run("returns error when env var is not a valid integer", func(t *testing.T) {
		t.Setenv("TEST_PORT_BAD", "not-a-number")

		_, err := config.GetInt("TEST_PORT_BAD", 3000)

		// We expect an error here — if there is no error, the test fails
		if err == nil {
			t.Error("expected an error for non-integer value, got nil")
		}
	})

	t.Run("returns error when env var is a float", func(t *testing.T) {
		// Ports and similar values must be whole integers.
		// "8080.5" should not silently become 8080.
		t.Setenv("TEST_PORT_FLOAT", "8080.5")

		_, err := config.GetInt("TEST_PORT_FLOAT", 3000)

		if err == nil {
			t.Error("expected an error for float value, got nil")
		}
	})
}

// TestGetBool covers boolean parsing.
func TestGetBool(t *testing.T) {
	t.Run("returns true for value 'true'", func(t *testing.T) {
		t.Setenv("TEST_FLAG", "true")

		result, err := config.GetBool("TEST_FLAG", false)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !result {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns true for value '1'", func(t *testing.T) {
		t.Setenv("TEST_FLAG", "1")

		result, err := config.GetBool("TEST_FLAG", false)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !result {
			t.Error("expected true, got false")
		}
	})

	t.Run("returns default when env var is not set", func(t *testing.T) {
		result, err := config.GetBool("TEST_FLAG_MISSING", true)

		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !result {
			t.Error("expected default true, got false")
		}
	})

	t.Run("returns error for invalid boolean value", func(t *testing.T) {
		t.Setenv("TEST_FLAG_BAD", "yes") // "yes" is not a valid Go bool

		_, err := config.GetBool("TEST_FLAG_BAD", false)

		if err == nil {
			t.Error("expected an error for invalid boolean, got nil")
		}
	})
}
