package config

import "testing"

func TestProcessMetricConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.ProcessMetrics || cfg.ProcessTopN != 10 {
		t.Fatalf("defaults = enabled %v, top N %d; want false, 10", cfg.ProcessMetrics, cfg.ProcessTopN)
	}

	for _, value := range []string{"0", "21", "nope"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("KANSHI_PROCESS_TOP_N", value)
			cfg := DefaultConfig()
			if err := LoadFromEnv(&cfg); err == nil {
				t.Fatal("expected invalid process top N to fail")
			}
		})
	}

	t.Setenv("KANSHI_PROCESS_METRICS", "true")
	t.Setenv("KANSHI_PROCESS_TOP_N", "20")
	cfg = DefaultConfig()
	if err := LoadFromEnv(&cfg); err != nil || !cfg.ProcessMetrics || cfg.ProcessTopN != 20 {
		t.Fatalf("process config = %+v, %v", cfg, err)
	}
}

func TestLoadFromEnvValidatesTLS(t *testing.T) {
	tests := []struct {
		name       string
		enabled    string
		caFile     string
		serverName string
	}{
		{name: "invalid boolean", enabled: "sometimes"},
		{name: "CA without TLS", enabled: "false", caFile: "/tmp/ca.pem"},
		{name: "server name without TLS", enabled: "false", serverName: "core.example.com"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("KANSHI_TLS", tt.enabled)
			t.Setenv("KANSHI_TLS_CA_FILE", tt.caFile)
			t.Setenv("KANSHI_TLS_SERVER_NAME", tt.serverName)
			cfg := DefaultConfig()
			if err := LoadFromEnv(&cfg); err == nil {
				t.Fatal("expected invalid TLS config to fail")
			}
		})
	}

	t.Run("valid", func(t *testing.T) {
		t.Setenv("KANSHI_TLS", "true")
		t.Setenv("KANSHI_TLS_CA_FILE", "/tmp/ca.pem")
		cfg := DefaultConfig()
		if err := LoadFromEnv(&cfg); err != nil {
			t.Fatalf("expected TLS config to load: %v", err)
		}
	})
}
