package constants

import "time"

// Server timeouts
const (
	DefaultReadTimeout     = 15 * time.Second
	DefaultWriteTimeout    = 15 * time.Second
	DefaultIdleTimeout     = 60 * time.Second
	DefaultShutdownTimeout = 10 * time.Second
	DefaultHeaderTimeout   = 5 * time.Second
	DefaultMaxHeaderBytes  = 1 << 20 // 1MB
)

// JWT settings
const (
	DefaultTokenExpiration = 24 * time.Hour
)

// Prediction points
const (
	ExactScorePoints    = 3
	CorrectResultPoints = 1
	NoPoints            = 0
)

// Initial values
const (
	InitialScore = 0
	InitialStats = 0
)
