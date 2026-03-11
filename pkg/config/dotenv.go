package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func LoadEnv(path string) (err error) {
	// #nosec G304 -- the caller controls the local .env path for development startup.
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return fmt.Errorf("open env file %q: %w", path, err)
	}
	defer func() {
		closeErr := file.Close()
		if err == nil && closeErr != nil {
			err = fmt.Errorf("close env file %q: %w", path, closeErr)
		}
	}()

	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++

		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		key, value, found := strings.Cut(line, "=")
		if !found {
			return fmt.Errorf("parse env file %q line %d: missing '='", path, lineNumber)
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if key == "" {
			return fmt.Errorf("parse env file %q line %d: empty key", path, lineNumber)
		}

		// Preserve real environment variables set by the shell or orchestrator.
		if _, exists := os.LookupEnv(key); exists {
			continue
		}

		if err := os.Setenv(key, trimQuotes(value)); err != nil {
			return fmt.Errorf("set env %q from %q line %d: %w", key, path, lineNumber, err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan env file %q: %w", path, err)
	}

	return nil
}

func trimQuotes(value string) string {
	if len(value) < 2 {
		return value
	}

	first, last := value[0], value[len(value)-1]
	if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
		return value[1 : len(value)-1]
	}

	return value
}
