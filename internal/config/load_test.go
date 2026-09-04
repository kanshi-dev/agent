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

func TestPprofTargetConfig(t *testing.T) {
	t.Setenv("KANSHI_PPROF_TARGETS", "checkout=http://127.0.0.1:6060,secure=https://go.internal:7443")
	t.Setenv("KANSHI_PPROF_DISCOVERY", "worker=http://localhost:6059-6061")
	cfg := DefaultConfig()
	if err := LoadFromEnv(&cfg); err != nil {
		t.Fatal(err)
	}
	if len(cfg.PprofTargets) != 2 || cfg.PprofTargets[0].URL != "http://127.0.0.1:6060" {
		t.Fatalf("targets = %+v", cfg.PprofTargets)
	}
	if len(cfg.PprofDiscovery) != 1 || cfg.PprofDiscovery[0].StartPort != 6059 || cfg.PprofDiscovery[0].EndPort != 6061 {
		t.Fatalf("discovery = %+v", cfg.PprofDiscovery)
	}
	if got := cfg.ProfileTargets(); len(got) != 3 || got[2].Name != "worker" || !got[2].Discovered {
		t.Fatalf("metadata = %+v", got)
	}
}

func TestPprofTargetConfigRejectsUnsafeInput(t *testing.T) {
	tests := map[string]struct{ targets, discovery string }{
		"credentials":   {targets: "bad=http://user:pass@localhost:6060"},
		"query":         {targets: "bad=http://localhost:6060?x=1"},
		"path":          {targets: "bad=http://localhost:6060/debug/pprof"},
		"scheme":        {targets: "bad=file://localhost:6060"},
		"duplicate":     {targets: "same=http://localhost:6060", discovery: "same=http://localhost:6061-6062"},
		"wide range":    {discovery: "bad=http://localhost:6000-6008"},
		"reverse range": {discovery: "bad=http://localhost:6061-6059"},
		"invalid name":  {targets: "bad name=http://localhost:6060"},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv("KANSHI_PPROF_TARGETS", tt.targets)
			t.Setenv("KANSHI_PPROF_DISCOVERY", tt.discovery)
			cfg := DefaultConfig()
			if err := LoadFromEnv(&cfg); err == nil {
				t.Fatal("expected invalid configuration")
			}
		})
	}
}
