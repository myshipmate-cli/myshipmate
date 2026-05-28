package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// EnvVar represents a single environment variable
type EnvVar struct {
	Key   string
	Value string
}

// ParseEnvFile reads a .env file and returns key-value pairs
func ParseEnvFile(path string) ([]EnvVar, error) {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	var vars []EnvVar
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first =
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove surrounding quotes
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		vars = append(vars, EnvVar{Key: key, Value: value})
	}

	return vars, scanner.Err()
}

// FindEnvFiles looks for .env files in the current directory
func FindEnvFiles() []string {
	patterns := []string{".env", ".env.local", ".env.production", ".env.development"}
	var found []string

	for _, p := range patterns {
		if _, err := os.Stat(p); err == nil {
			found = append(found, p)
		}
	}

	return found
}

// LoadAllEnvVars loads all env vars from found .env files
func LoadAllEnvVars() ([]EnvVar, error) {
	files := FindEnvFiles()
	if len(files) == 0 {
		return nil, nil
	}

	// Use a map to deduplicate (later files override earlier ones)
	envMap := make(map[string]string)
	var order []string

	for _, f := range files {
		vars, err := ParseEnvFile(f)
		if err != nil {
			return nil, fmt.Errorf("error reading %s: %w", f, err)
		}
		for _, v := range vars {
			if _, exists := envMap[v.Key]; !exists {
				order = append(order, v.Key)
			}
			envMap[v.Key] = v.Value
		}
	}

	var result []EnvVar
	for _, k := range order {
		result = append(result, EnvVar{Key: k, Value: envMap[k]})
	}

	return result, nil
}

// DisplayEnvSummary shows a summary of detected env vars (values masked)
func DisplayEnvSummary(vars []EnvVar) {
	if len(vars) == 0 {
		return
	}

	fmt.Printf("   📋 Found %d environment variables:\n", len(vars))
	for _, v := range vars {
		masked := maskValue(v.Value)
		fmt.Printf("      • %s = %s\n", v.Key, masked)
	}
}

// maskValue hides most of a value for display
func maskValue(value string) string {
	if len(value) <= 4 {
		return "****"
	}
	return value[:2] + strings.Repeat("*", min(len(value)-4, 8)) + value[len(value)-2:]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
