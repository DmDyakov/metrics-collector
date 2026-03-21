package config

import (
	"flag"

	"github.com/caarlos0/env/v11"
)

type AgentConfig struct {
	PollInterval   int    `env:"POLL_INTERVAL"`
	ReportInterval int    `env:"REPORT_INTERVAL"`
	ServerBaseURL  string `env:"ADDRESS"`
}

type ServerConfig struct {
	ServerBaseURL string `env:"ADDRESS"`
}

const (
	defaultServerBaseURL  = "localhost:8080"
	defaultPollInterval   = 2
	defaultReportInterval = 10
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
		cfg.ServerBaseURL = serverBaseURLFlag
	}

	if cfg.PollInterval == 0 {
		cfg.PollInterval = pollIntervalFlag
	}

	if cfg.ReportInterval == 0 {
		cfg.ReportInterval = reportIntervalFlag
	}

	return cfg, nil
}

func NewServerConfig(args []string) (*ServerConfig, error) {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)

	var serverBaseURLFlag string

	fs.StringVar(&serverBaseURLFlag, "a", defaultServerBaseURL, "address and port to run server")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	cfg := &ServerConfig{
		ServerBaseURL: serverBaseURLFlag,
	}

	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	if cfg.ServerBaseURL == "" {
		cfg.ServerBaseURL = serverBaseURLFlag
	}

	return cfg, nil
}
