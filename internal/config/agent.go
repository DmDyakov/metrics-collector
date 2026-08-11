package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
)

type AgentConfig struct {
	PollInterval    int           `env:"POLL_INTERVAL" json:"poll_interval"`
	ReportInterval  int           `env:"REPORT_INTERVAL" json:"report_interval"`
	HTTPAddress     string        `env:"ADDRESS" json:"address"`
	GRPCAddress     string        `env:"GRPC_ADDRESS" json:"grpc_address"`
	SecretKey       string        `env:"KEY" json:"key"`
	PublicCryptoKey string        `env:"CRYPTO_KEY" json:"crypto_key"`
	RateLimit       int           `env:"RATE_LIMIT" json:"rate_limit"`
	AgentIP         string        `env:"AGENT_IP" json:"agent_ip"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" json:"shutdown_timeout"`
}

const (
	defaultServerBaseURL        = "localhost:8080"
	defaultPollInterval         = 2
	defaultReportInterval       = 10
	defaultRateLimit            = 2
	defaultAgentShutdownTimeout = 10 * time.Second
)

func NewAgentConfig(args []string) (*AgentConfig, error) {
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)

	var configPath string
	fs.StringVar(&configPath, "c", "", "path to JSON config file")
	fs.StringVar(&configPath, "config", "", "path to JSON config file")

	addr := fs.String("a", defaultServerBaseURL, "address and port")
	poll := fs.Int("p", defaultPollInterval, "poll interval")
	report := fs.Int("r", defaultReportInterval, "report interval")
	key := fs.String("k", "", "secret key")
	limit := fs.Int("l", defaultRateLimit, "rate limit")
	crypto := fs.String("crypto-key", "", "path to public key")
	grpcAddr := fs.String("grpc-addr", "", "gRPC server address")
	shutTimeout := fs.Duration("s", defaultAgentShutdownTimeout, "shutdown timeout")

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}

	cfg := &AgentConfig{
		HTTPAddress:     *addr,
		PollInterval:    *poll,
		ReportInterval:  *report,
		SecretKey:       *key,
		RateLimit:       *limit,
		PublicCryptoKey: *crypto,
		GRPCAddress:     *grpcAddr,
		ShutdownTimeout: *shutTimeout,
	}

	if configPath == "" {
		configPath = os.Getenv("CONFIG")
	}
	if err := loadJSONFile(configPath, cfg); err != nil {
		return nil, fmt.Errorf("failed to load JSON config: %w", err)
	}
	if err := loadDotEnv(); err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}
	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse env: %w", err)
	}

	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			cfg.HTTPAddress = *addr
		case "p":
			cfg.PollInterval = *poll
		case "r":
			cfg.ReportInterval = *report
		case "k":
			cfg.SecretKey = *key
		case "l":
			cfg.RateLimit = *limit
		case "crypto-key":
			cfg.PublicCryptoKey = *crypto
		case "grpc-addr":
			cfg.GRPCAddress = *grpcAddr
		}
	})

	if cfg.HTTPAddress == "" {
		return nil, errors.New("server URL can not be empty")
	}
	if cfg.AgentIP == "" {
		return nil, errors.New("agent ip can not be empty")
	}
	if cfg.PollInterval <= 0 {
		return nil, errors.New("poll interval must be positive")
	}
	if cfg.ReportInterval <= 0 {
		return nil, errors.New("report interval must be positive")
	}
	if cfg.RateLimit <= 0 {
		cfg.RateLimit = defaultRateLimit
	}
	if cfg.ReportInterval < cfg.PollInterval {
		return nil, fmt.Errorf("report interval (%d) must be >= poll interval (%d)", cfg.ReportInterval, cfg.PollInterval)
	}

	return cfg, nil
}
