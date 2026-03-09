package secrets

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ParseKeyValues parses a slice of KEY=VALUE strings into a map.
// Values are cleaned (trimmed and unquoted) consistently.
func ParseKeyValues(entries []string) (map[string]string, error) {
	result := make(map[string]string, len(entries))
	for _, entry := range entries {
		key, value, err := parseEntry(entry)
		if err != nil {
			return nil, fmt.Errorf("invalid secret %q: %w", entry, err)
		}
		result[key] = value
	}
	return result, nil
}

// ParseEnvFile reads a .env file and returns secrets as a map.
// Returns an empty map if the file does not exist.
func ParseEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, fmt.Errorf("failed to open env file: %w", err)
	}
	defer file.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(file)
	lineNumber := 0

	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || line[0] == '#' {
			continue
		}

		if strings.HasPrefix(line, "export ") {
			line = strings.TrimSpace(strings.TrimPrefix(line, "export "))
		}

		key, value, parseErr := parseEntry(line)
		if parseErr != nil {
			return nil, fmt.Errorf("env file line %d: %w", lineNumber, parseErr)
		}

		result[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read env file: %w", err)
	}

	return result, nil
}

// Merge combines multiple secret maps. Later maps override earlier ones.
func Merge(maps ...map[string]string) map[string]string {
	size := 0
	for _, m := range maps {
		size += len(m)
	}
	merged := make(map[string]string, size)
	for _, m := range maps {
		for k, v := range m {
			merged[k] = v
		}
	}
	return merged
}

// ParseOSEnv collects all current OS environment variables into a map.
func ParseOSEnv() map[string]string {
	result := make(map[string]string)
	for _, entry := range os.Environ() {
		if k, v, ok := strings.Cut(entry, "="); ok && k != "" {
			result[k] = v
		}
	}
	return result
}

func parseEntry(entry string) (string, string, error) {
	parts := strings.SplitN(entry, "=", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected KEY=VALUE")
	}

	key := strings.TrimSpace(parts[0])
	if key == "" {
		return "", "", fmt.Errorf("key cannot be empty")
	}

	value := strings.TrimSpace(parts[1])

	// Strip inline comments (only when value is unquoted)
	if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') {
		quote := value[0]
		if end := strings.LastIndexByte(value, quote); end > 0 {
			value = value[1:end]
		}
	} else if idx := strings.Index(value, " #"); idx >= 0 {
		value = strings.TrimSpace(value[:idx])
	}

	return key, value, nil
}
