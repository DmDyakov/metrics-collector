package config

import "time"

type Config struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	ServerBaseURL  string
}

func NewConfig() *Config {
	return &Config{
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		ServerBaseURL:  "localhost:8080",
	}
}
