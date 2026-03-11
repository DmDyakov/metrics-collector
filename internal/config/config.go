package config

import (
	"flag"
	"time"
)

type AgentConfig struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	ServerBaseURL  string
}

type ServerConfig struct {
	ServerBaseURL string
}

var (
	flagRunAddr        string
	flagReportInterval string
	flagPollInterval   string
)

func parseFlags() {
	const (
		defaultServerBaseURL  = "localhost:8080"
		defaultPollInterval   = "2s"
		defaultReportInterval = "10s"
	)

	flag.StringVar(&flagRunAddr, "a", defaultServerBaseURL, "address and port to run server")
	flag.StringVar(&flagPollInterval, "p", defaultPollInterval, "poll interval")
	flag.StringVar(&flagReportInterval, "r", defaultReportInterval, "report interval")
	flag.Parse()
}

func NewAgentConfig() (*AgentConfig, error) {
	parseFlags()
	pollDur, err := time.ParseDuration(flagPollInterval)
	if err != nil {
		return nil, err
	}

	reportDur, err := time.ParseDuration(flagReportInterval)
	if err != nil {
		return nil, err
	}

	return &AgentConfig{
		PollInterval:   pollDur,
		ReportInterval: reportDur,
		ServerBaseURL:  flagRunAddr,
	}, nil
}

func NewServerConfig() *ServerConfig {
	parseFlags()
	return &ServerConfig{
		ServerBaseURL: flagRunAddr,
	}
}
