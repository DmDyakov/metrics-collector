package config

import (
	"flag"
	"strconv"
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
	flagServerBaseURL  string
	flagReportInterval string
	flagPollInterval   string
)

func parseFlags() {
	const (
		defaultServerBaseURL  = "localhost:8080"
		defaultPollInterval   = "2"
		defaultReportInterval = "10"
	)

	flag.StringVar(&flagServerBaseURL, "a", defaultServerBaseURL, "address and port to run server")
	flag.StringVar(&flagPollInterval, "p", defaultPollInterval, "poll interval")
	flag.StringVar(&flagReportInterval, "r", defaultReportInterval, "report interval")
	flag.Parse()
}

func NewAgentConfig() (*AgentConfig, error) {
	parseFlags()

	pollSec, err := strconv.ParseInt(flagPollInterval, 10, 64)
	if err != nil {
		return nil, err
	}

	reportSec, err := strconv.ParseInt(flagReportInterval, 10, 64)
	if err != nil {
		return nil, err
	}

	pollDur := time.Duration(pollSec) * time.Second
	reportDur := time.Duration(reportSec) * time.Second

	return &AgentConfig{
		PollInterval:   pollDur,
		ReportInterval: reportDur,
		ServerBaseURL:  flagServerBaseURL,
	}, nil
}

func NewServerConfig() *ServerConfig {
	parseFlags()
	return &ServerConfig{
		ServerBaseURL: flagServerBaseURL,
	}
}
