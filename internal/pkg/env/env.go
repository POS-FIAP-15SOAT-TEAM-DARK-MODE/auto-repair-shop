package env

import (
	"os"
	"strconv"
)

// GetString retrieves the value of an environment variable as a string.
// If the variable is not set, it returns the provided default value.
func GetString(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetInt retrieves the value of an environment variable as an integer.
// If the variable is not set or cannot be parsed, it returns the provided default value.
func GetInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}

	return intValue
}
