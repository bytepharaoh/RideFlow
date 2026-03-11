// Package config provides helpers for reading configuration
// from environment variables.
package config

import (
	"fmt"
	"os"
	"strconv"
)

func GetString(key, defaultValue string) string {
	// os.Getenv returns "" for both "not set" and "set to empty string".
	// We treat both the same way: use the default.

	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
func GetInt(key string, defaultValue int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("config: %s must be an integer, got %q: %w", key, value, err)

	}
	return parsed, nil
}
func GetBool(key string, defaultValue bool) (bool, error) {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("config: %s must be a boolean, got %q: %w", key, value, err)
	}
	return parsed, nil
}
