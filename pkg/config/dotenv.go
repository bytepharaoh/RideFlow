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
		if shouldSkipEnvLine(line) {
			continue
		}

		if err := setEnvFromLine(path, lineNumber, line); err != nil {
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan env file %q: %w", path, err)
	}

	return nil
}

func shouldSkipEnvLine(line string) bool {
	return line == "" || strings.HasPrefix(line, "#")
}

func setEnvFromLine(path string, lineNumber int, line string) error {
	key, value, err := parseEnvLine(path, lineNumber, line)
	if err != nil {
		return err
	}

	// Preserve real environment variables set by the shell or orchestrator.
	if _, exists := os.LookupEnv(key); exists {
		return nil
	}

	if err := os.Setenv(key, trimQuotes(value)); err != nil {
		return fmt.Errorf("set env %q from %q line %d: %w", key, path, lineNumber, err)
	}

	return nil
}

func parseEnvLine(path string, lineNumber int, line string) (string, string, error) {
	key, value, found := strings.Cut(line, "=")
	if !found {
		return "", "", fmt.Errorf("parse env file %q line %d: missing '='", path, lineNumber)
	}

	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if key == "" {
		return "", "", fmt.Errorf("parse env file %q line %d: empty key", path, lineNumber)
	}

	return key, value, nil
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
