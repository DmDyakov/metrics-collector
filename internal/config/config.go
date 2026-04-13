package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v11"
)

type AgentConfig struct {
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	ServerBaseURL  string `env:"ADDRESS"`
}

type ServerConfig struct {
	ServerBaseURL   string `env:"ADDRESS"`
	StoreInterval   int    `env:"STORE_INTERVAL"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	Restore         bool   `env:"RESTORE"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

const (
	defaultServerBaseURL   = "localhost:8080"
	defaultPollInterval    = 2
	defaultReportInterval  = 10
	defaultStoreInterval   = 300
	defaultFileStoragePath = "store.json"
	defaultRestore         = false
	defaultDatabaseDSN     = "postgres://DmtiryDyakov:@localhost:5432/metrics?sslmode=disable"
)

func NewAgentConfig(args []string) (*AgentConfig, error) {
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)

	cfg := &AgentConfig{}

	fs.StringVar(&cfg.ServerBaseURL, "a", defaultServerBaseURL, "address and port to run server")
	fs.IntVar(&cfg.PollInterval, "p", defaultPollInterval, "poll interval")
	fs.IntVar(&cfg.ReportInterval, "r", defaultReportInterval, "report interval")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	if cfg.ServerBaseURL == "" {
		return nil, fmt.Errorf("server URL can not be empty")
	}

	if cfg.PollInterval <= 0 {
		return nil, fmt.Errorf("poll interval must be positive")
	}

	if cfg.ReportInterval <= 0 {
		return nil, fmt.Errorf("report interval must be positive")
	}

	if cfg.ReportInterval < cfg.PollInterval {
		return nil, fmt.Errorf("report interval (%d) must be greater than or equal to poll interval (%d)", cfg.ReportInterval, cfg.PollInterval)
	}

	return cfg, nil
}

func NewServerConfig(args []string) (*ServerConfig, error) {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)

	cfg := &ServerConfig{}

	fs.StringVar(&cfg.ServerBaseURL, "a", defaultServerBaseURL, "address and port to run server")
	fs.IntVar(&cfg.StoreInterval, "i", defaultStoreInterval, "store interval")
	fs.StringVar(&cfg.FileStoragePath, "f", defaultFileStoragePath, "file storage path")
	fs.BoolVar(&cfg.Restore, "r", defaultRestore, "restore")
	fs.StringVar(&cfg.DatabaseDSN, "d", defaultDatabaseDSN, "database DSN")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	if cfg.ServerBaseURL == "" {
		return nil, fmt.Errorf("server URL can not be empty")
	}

	if cfg.StoreInterval < 0 {
		return nil, fmt.Errorf("store interval must be non-negative")
	}

	if cfg.FileStoragePath == "" {
		return nil, fmt.Errorf("file storage path can not be empty")
	}

	if cfg.DatabaseDSN == "" {
		return nil, fmt.Errorf("database DSN can not be empty")
	}

	return cfg, nil
}
