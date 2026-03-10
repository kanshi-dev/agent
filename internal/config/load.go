package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

func LoadFromEnv(c *Config) {
	if v := os.Getenv("KANSHI_CORE_ADDR"); v != "" {
		c.CoreAddr = v
	}

	if v := os.Getenv("KANSHI_API_KEY"); v != "" {
		c.APIKey = v
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
}
