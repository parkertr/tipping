package utils

import "time"

// GenerateID generates a unique ID based on the current timestamp
func GenerateID() string {
	return time.Now().Format("20060102150405.000000000")
}
