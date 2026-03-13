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

const (
	defaultServerBaseURL  = "localhost:8080"
	defaultPollInterval   = "2"
	defaultReportInterval = "10"
)

func NewAgentConfig() (*AgentConfig, error) {
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)

	var (
		serverBaseURL  string
		reportInterval string
		pollInterval   string
	)

	fs.StringVar(&serverBaseURL, "a", defaultServerBaseURL, "address and port to run server")
	fs.StringVar(&pollInterval, "p", defaultPollInterval, "poll interval")
	fs.StringVar(&reportInterval, "r", defaultReportInterval, "report interval")

	if err := fs.Parse(flag.Args()); err != nil {
		return nil, err
	}

	pollSec, err := strconv.ParseInt(pollInterval, 10, 64)
	if err != nil {
		return nil, err
	}

	reportSec, err := strconv.ParseInt(reportInterval, 10, 64)
	if err != nil {
		return nil, err
	}

	pollDur := time.Duration(pollSec) * time.Second
	reportDur := time.Duration(reportSec) * time.Second

	return &AgentConfig{
		PollInterval:   pollDur,
		ReportInterval: reportDur,
		ServerBaseURL:  serverBaseURL,
	}, nil
}

func NewServerConfig() (*ServerConfig, error) {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)

	var serverBaseURL string

	fs.StringVar(&serverBaseURL, "a", defaultServerBaseURL, "address and port to run server")

	if err := fs.Parse(flag.Args()); err != nil {
		return nil, err
	}

	return &ServerConfig{
		ServerBaseURL: serverBaseURL,
	}, nil
}
