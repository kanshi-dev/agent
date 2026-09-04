package config

import (
	"fmt"
	"regexp"
	"time"
)

var profileTargetName = regexp.MustCompile(`^[A-Za-z0-9._-]{1,64}$`)

type PprofTarget struct {
	Name string
	URL  string
}

type PprofDiscovery struct {
	Name      string
	Scheme    string
	Host      string
	StartPort int
	EndPort   int
}

type ProfileTargetMetadata struct {
	Name       string
	Discovered bool
}

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
	// PprofTargets are explicit operator-approved Go pprof endpoints.
	PprofTargets []PprofTarget
	// PprofDiscovery contains operator-approved single-host port ranges.
	PprofDiscovery []PprofDiscovery
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
	if len(c.PprofTargets)+len(c.PprofDiscovery) > 20 {
		return fmt.Errorf("pprof targets cannot exceed 20 names")
	}
	seen := make(map[string]struct{}, len(c.PprofTargets)+len(c.PprofDiscovery))
	for _, target := range c.ProfileTargets() {
		if !profileTargetName.MatchString(target.Name) {
			return fmt.Errorf("pprof target name %q must match [A-Za-z0-9._-]{1,64}", target.Name)
		}
		if _, ok := seen[target.Name]; ok {
			return fmt.Errorf("pprof target name %q is duplicated", target.Name)
		}
		seen[target.Name] = struct{}{}
	}
	return nil
}

func (c Config) ProfileTargets() []ProfileTargetMetadata {
	targets := make([]ProfileTargetMetadata, 0, len(c.PprofTargets)+len(c.PprofDiscovery))
	for _, target := range c.PprofTargets {
		targets = append(targets, ProfileTargetMetadata{Name: target.Name})
	}
	for _, target := range c.PprofDiscovery {
		targets = append(targets, ProfileTargetMetadata{Name: target.Name, Discovered: true})
	}
	return targets
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
