// Package envfile loads a .env file (KEY=value) into the process environment.
package envfile

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Load reads the file at path. Variables that are already set (shell, Docker)
// keep their existing value. A missing file is not an error.
func Load(path string) error {
	f, err := os.Open(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")

		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("%s line %d: missing '='", path, lineNumber)
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		if isQuoted(value) {
			value = value[1 : len(value)-1]
		} else if i := strings.Index(value, " #"); i >= 0 {
			value = strings.TrimSpace(value[:i])
		}

		if _, alreadySet := os.LookupEnv(key); !alreadySet {
			os.Setenv(key, value)
		}
	}
	return scanner.Err()
}

func isQuoted(s string) bool {
	if len(s) < 2 {
		return false
	}
	quote := s[0]
	return (quote == '"' || quote == '\'') && s[len(s)-1] == quote
}

// String returns the value of key, or fallback if it is unset or empty.
func String(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

// Int returns the value of key parsed as an integer, or fallback.
func Int(key string, fallback int) int {
	if v, err := strconv.Atoi(String(key, "")); err == nil {
		return v
	}
	return fallback
}

// Bool accepts true/false, 1/0, yes/no.
func Bool(key string, fallback bool) bool {
	switch strings.ToLower(String(key, "")) {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	}
	return fallback
}

// Missing returns which of the given keys are unset or empty.
func Missing(keys ...string) []string {
	var missing []string
	for _, k := range keys {
		if String(k, "") == "" {
			missing = append(missing, k)
		}
	}
	return missing
}
