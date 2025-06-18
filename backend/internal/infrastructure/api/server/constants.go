package server

import "time"

const (
	// DefaultReadTimeout is the default timeout for reading the entire request
	DefaultReadTimeout = 5 * time.Second

	// DefaultWriteTimeout is the default timeout for writing the response
	DefaultWriteTimeout = 10 * time.Second

	// DefaultIdleTimeout is the default timeout for keeping idle connections
	DefaultIdleTimeout = 120 * time.Second
)
