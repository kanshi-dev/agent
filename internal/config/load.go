package config

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// LoadFromEnv populates the Config from environment variables, using KANSHI_ prefix.
func LoadFromEnv(c *Config) error {
	if v := os.Getenv("KANSHI_CORE_ADDR"); v != "" {
		c.CoreAddr = v
	}

	if v := os.Getenv("KANSHI_API_KEY"); v != "" {
		c.APIKey = v
	}

	if v := os.Getenv("KANSHI_TLS"); v != "" {
		enabled, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("KANSHI_TLS must be true or false: %w", err)
		}
		c.TLS = enabled
	}

	if v := os.Getenv("KANSHI_TLS_CA_FILE"); v != "" {
		c.TLSCAFile = v
	}
	if v := os.Getenv("KANSHI_TLS_SERVER_NAME"); v != "" {
		c.TLSServerName = v
	}

	if v := os.Getenv("KANSHI_LOG_LEVEL"); v != "" {
		c.LogLevel = strings.ToLower(v)
	}

	if v := os.Getenv("KANSHI_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			c.Interval = d
		}
	}

	if v := os.Getenv("KANSHI_BATCH_MAX"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.BatchMax = n
		}
	}

	if v := os.Getenv("KANSHI_FLUSH_EVERY"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			c.FlushEvery = d
		}
	}

	if v := os.Getenv("KANSHI_HOST_TAGS"); v != "" {
		c.HostTags = strings.Split(v, ",")
	}

	if v := os.Getenv("KANSHI_PROCESS_METRICS"); v != "" {
		enabled, err := strconv.ParseBool(v)
		if err != nil {
			return fmt.Errorf("KANSHI_PROCESS_METRICS must be true or false: %w", err)
		}
		c.ProcessMetrics = enabled
	}

	if v := os.Getenv("KANSHI_PROCESS_TOP_N"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return fmt.Errorf("KANSHI_PROCESS_TOP_N must be an integer: %w", err)
		}
		c.ProcessTopN = n
	}
	var err error
	if c.PprofTargets, err = parsePprofTargets(os.Getenv("KANSHI_PPROF_TARGETS")); err != nil {
		return fmt.Errorf("KANSHI_PPROF_TARGETS: %w", err)
	}
	if c.PprofDiscovery, err = parsePprofDiscovery(os.Getenv("KANSHI_PPROF_DISCOVERY")); err != nil {
		return fmt.Errorf("KANSHI_PPROF_DISCOVERY: %w", err)
	}

	return c.Validate()
}

func entries(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	return strings.Split(raw, ",")
}

func parsePprofTargets(raw string) ([]PprofTarget, error) {
	var targets []PprofTarget
	for _, entry := range entries(raw) {
		name, address, ok := strings.Cut(strings.TrimSpace(entry), "=")
		if !ok {
			return nil, fmt.Errorf("entry %q must be name=URL", entry)
		}
		u, err := url.Parse(address)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" || u.Port() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || (u.Path != "" && u.Path != "/") {
			return nil, fmt.Errorf("target %q must use an HTTP or HTTPS host and port without credentials, path, query, or fragment", name)
		}
		port, err := strconv.Atoi(u.Port())
		if err != nil || port < 1 || port > 65535 {
			return nil, fmt.Errorf("target %q has an invalid port", name)
		}
		u.Path = ""
		targets = append(targets, PprofTarget{Name: name, URL: u.String()})
	}
	return targets, nil
}

func parsePprofDiscovery(raw string) ([]PprofDiscovery, error) {
	var ranges []PprofDiscovery
	for _, entry := range entries(raw) {
		name, address, ok := strings.Cut(strings.TrimSpace(entry), "=")
		if !ok {
			return nil, fmt.Errorf("entry %q must be name=URL", entry)
		}
		dash := strings.LastIndex(address, "-")
		colon := strings.LastIndex(address, ":")
		if dash <= colon || colon < 0 {
			return nil, fmt.Errorf("discovery %q must end in start-end ports", name)
		}
		start, err1 := strconv.Atoi(address[colon+1 : dash])
		end, err2 := strconv.Atoi(address[dash+1:])
		base, err := url.Parse(address[:colon] + ":" + strconv.Itoa(start))
		if err != nil || err1 != nil || err2 != nil || (base.Scheme != "http" && base.Scheme != "https") || base.Hostname() == "" || base.User != nil || base.RawQuery != "" || base.Fragment != "" || base.Path != "" || start < 1 || end < start || end > 65535 || end-start+1 > 8 {
			return nil, fmt.Errorf("discovery %q must use one HTTP or HTTPS host and at most eight valid ports", name)
		}
		if net.ParseIP(base.Hostname()) == nil && strings.Contains(base.Hostname(), ":") {
			return nil, fmt.Errorf("discovery %q has an invalid host", name)
		}
		ranges = append(ranges, PprofDiscovery{Name: name, Scheme: base.Scheme, Host: base.Hostname(), StartPort: start, EndPort: end})
	}
	return ranges, nil
}
