package config

import (
	"fmt"
	"time"
)

type Config struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	Host           string
	Port           string
}

func NewConfig() *Config {
	return &Config{
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		Host:           "localhost",
		Port:           "8080",
	}
}

func (cfg *Config) ServerBaseURL() string {
	return fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
}
