package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

func LoadFromEnv(c *Config) {
	c.CoreAddr = os.Getenv("KANSHI_CORE_ADDR")
	c.APIKey = os.Getenv("KANSHI_API_KEY")
	d, _ := time.ParseDuration(os.Getenv("KANSHI_INTERVAL"))
	c.Interval = d
	n, _ := strconv.Atoi(os.Getenv("KANSHI_BATCH_MAX"))
	c.BatchMax = n
	f, _ := time.ParseDuration(os.Getenv("KANSHI_FLUSH_EVERY"))
	c.FlushEvery = f
	c.HostTags = strings.Split(os.Getenv("KANSHI_HOST_TAGS"), ",")
}
