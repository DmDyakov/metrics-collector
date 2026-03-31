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
}

const (
	defaultServerBaseURL   = "localhost:8080"
	defaultPollInterval    = 2
	defaultReportInterval  = 10
	defaultStoreInterval   = 300
	defaultFileStoragePath = "store.json"
	defaultRestore         = false
)

func NewAgentConfig(args []string) (*AgentConfig, error) {
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)

	var (
		serverBaseURLFlag  string
		pollIntervalFlag   int
		reportIntervalFlag int
	)

	fs.StringVar(&serverBaseURLFlag, "a", defaultServerBaseURL, "address and port to run server")
	fs.IntVar(&pollIntervalFlag, "p", defaultPollInterval, "poll interval")
	fs.IntVar(&reportIntervalFlag, "r", defaultReportInterval, "report interval")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	cfg := &AgentConfig{
		ServerBaseURL:  serverBaseURLFlag,
		PollInterval:   pollIntervalFlag,
		ReportInterval: reportIntervalFlag,
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

	var (
		serverBaseURLFlag   string
		storeIntervalFlag   int
		fileStoragePathFlag string
		restoreFlag         bool
	)

	fs.StringVar(&serverBaseURLFlag, "a", defaultServerBaseURL, "address and port to run server")
	fs.IntVar(&storeIntervalFlag, "i", defaultStoreInterval, "store interval")
	fs.StringVar(&fileStoragePathFlag, "f", defaultFileStoragePath, "file storage path")
	fs.BoolVar(&restoreFlag, "r", defaultRestore, "restore")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	cfg := &ServerConfig{
		ServerBaseURL:   serverBaseURLFlag,
		StoreInterval:   storeIntervalFlag,
		FileStoragePath: fileStoragePathFlag,
		Restore:         restoreFlag,
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

	return cfg, nil
}
