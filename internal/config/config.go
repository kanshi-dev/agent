package config

import (
	"fmt"
	"time"
)

// Config represents the agent's configuration parameters.
type Config struct {
	// CoreAddr is the gRPC address of the Kanshi core service (e.g., "127.0.0.1:50051").
	CoreAddr string
	// APIKey is used for authentication with the core service.
	APIKey string
	// TLS enables TLS for the connection to Core.
	TLS bool
	// TLSCAFile optionally adds a PEM-encoded CA certificate.
	TLSCAFile string
	// TLSServerName optionally overrides the server name verified by TLS.
	TLSServerName string
	// LogLevel is the logging level (e.g., "info", "debug").
	LogLevel string
	// Interval defines how often the agent collects system metrics.
	Interval time.Duration
	// BatchMax is the maximum number of points to batch before flushing.
	BatchMax int
	// FlushEvery is the maximum time to wait before flushing regardless of batch size.
	FlushEvery time.Duration
	// HostTags are optional tags appended to all metrics collected by this host.
	HostTags []string
	// ProcessMetrics enables process count, CPU, and resident-memory metrics.
	ProcessMetrics bool
	// ProcessTopN limits the independent CPU and resident-memory rankings.
	ProcessTopN int
}

// Validate rejects configuration combinations that cannot be used safely.
func (c Config) Validate() error {
	if !c.TLS && c.TLSCAFile != "" {
		return fmt.Errorf("KANSHI_TLS_CA_FILE requires KANSHI_TLS=true")
	}
	if !c.TLS && c.TLSServerName != "" {
		return fmt.Errorf("KANSHI_TLS_SERVER_NAME requires KANSHI_TLS=true")
	}
	if c.ProcessTopN < 1 || c.ProcessTopN > 20 {
		return fmt.Errorf("KANSHI_PROCESS_TOP_N must be between 1 and 20")
	}
	return nil
}

// DefaultConfig returns a Config with sensible default values.
func DefaultConfig() Config {
	return Config{
		CoreAddr:    "127.0.0.1:50051",
		APIKey:      "",
		LogLevel:    "info",
		Interval:    5 * time.Second,
		BatchMax:    100,
		FlushEvery:  10 * time.Second,
		HostTags:    []string{},
		ProcessTopN: 10,
	}
}
