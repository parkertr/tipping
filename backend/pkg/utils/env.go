package utils

import (
	"os"
	"time"
)

// GetEnvOrDefault returns the value of the environment variable or a default value
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvOrDefaultInt returns the value of the environment variable as an int or a default value
func GetEnvOrDefaultInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := time.ParseDuration(value); err == nil {
			return int(intValue)
		}
	}
	return defaultValue
}

// GetEnvOrDefaultDuration returns the value of the environment variable as a duration or a default value
func GetEnvOrDefaultDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
