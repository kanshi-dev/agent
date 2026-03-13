package config

import "time"

// Config represents the agent's configuration parameters.
type Config struct {
	// CoreAddr is the gRPC address of the Kanshi core service (e.g., "127.0.0.1:50051").
	CoreAddr string
	// APIKey is used for authentication with the core service (currently unused).
	APIKey string
	// Interval defines how often the agent collects system metrics.
	Interval time.Duration
	// BatchMax is the maximum number of points to batch before flushing.
	BatchMax int
	// FlushEvery is the maximum time to wait before flushing regardless of batch size.
	FlushEvery time.Duration
	// HostTags are optional tags appended to all metrics collected by this host.
	HostTags []string
}

// DefaultConfig returns a Config with sensible default values.
func DefaultConfig() Config {
	return Config{
		CoreAddr:   "127.0.0.1:50051",
		APIKey:     "",
		Interval:   5 * time.Second,
		BatchMax:   100,
		FlushEvery: 10 * time.Second,
		HostTags:   []string{},
	}
}
