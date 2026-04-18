package config

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
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
	defaultStoreInterval   = 20
	defaultFileStoragePath = ""
	defaultRestore         = false
	defaultDatabaseDSN     = ""
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

	loadDotEnv()
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	if cfg.ServerBaseURL == "" {
		return nil, errors.New("server URL can not be empty")
	}

	if cfg.PollInterval <= 0 {
		return nil, errors.New("poll interval must be positive")
	}

	if cfg.ReportInterval <= 0 {
		return nil, errors.New("report interval must be positive")
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

	loadDotEnv()
	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	if cfg.ServerBaseURL == "" {
		return nil, errors.New("server URL can not be empty")
	}

	if cfg.StoreInterval < 0 {
		return nil, errors.New("store interval must be non-negative")
	}

	return cfg, nil
}

func loadDotEnv() {
	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			log.Printf("Warning: could not load .env file: %v", err)
		}
	}
}
