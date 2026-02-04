package config

import "time"

type Config struct {
	CoreAddr   string
	APIKey     string
	Interval   time.Duration
	BatchMax   int
	FlushEvery time.Duration
	HostTags   []string
}

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
