package utils

import (
	"os"
	"strconv"
)

// StringEnvOrDefault returns the value of the environment variable named by the key, or defaultValue if the environment variable is empty or not set.
func StringEnvOrDefault(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		return value
	}
	return defaultValue
}

// IntEnvOrDefault returns the value of the environment variable named by the key, or defaultValue if the environment variable is empty or not set.
func IntEnvOrDefault(key string, defaultValue int) int {
	if value, ok := os.LookupEnv(key); ok && value != "" {
		number, ok := strconv.Atoi(value)
		if ok == nil {
			return number
		}
	}
	return defaultValue
}
