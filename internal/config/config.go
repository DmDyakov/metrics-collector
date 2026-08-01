// Package config предоставляет функции для загрузки и парсинга конфигурации
// из переменных окружения и флагов командной строки.
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type AgentConfig struct {
	PollInterval    int    `env:"POLL_INTERVAL" json:"poll_interval"`
	ReportInterval  int    `env:"REPORT_INTERVAL" json:"report_interval"`
	ServerBaseURL   string `env:"ADDRESS" json:"address"`
	SecretKey       string `env:"KEY" json:"key"`
	PublicCryptoKey string `env:"CRYPTO_KEY" json:"crypto_key"`
	RateLimit       int    `env:"RATE_LIMIT" json:"rate_limit"`
	AgentIP         string `env:"AGENT_IP" json:"agent_ip"`
}

type ServerConfig struct {
	ServerBaseURL    string        `env:"ADDRESS" json:"address"`
	StoreInterval    int           `env:"STORE_INTERVAL" json:"store_interval"`
	FileStoragePath  string        `env:"FILE_STORAGE_PATH" json:"store_file"`
	Restore          bool          `env:"RESTORE" json:"restore"`
	DatabaseDSN      string        `env:"DATABASE_DSN" json:"database_dsn"`
	SecretKey        string        `env:"KEY" json:"key"`
	PrivateCryptoKey string        `env:"CRYPTO_KEY" json:"crypto_key"`
	RequestTimeout   time.Duration `env:"REQ_TIMEOUT" json:"request_timeout"`
	ShutdownTimeout  time.Duration `env:"SHUTDOWN_TIMEOUT" json:"shutdown_timeout"`
	AuditFile        string        `env:"AUDIT_FILE" json:"audit_file"`
	AuditURL         string        `env:"AUDIT_URL" json:"audit_url"`
	PprofAddr        string        `env:"PPROF_ADDR" json:"pprof_addr"`
	TrustedSubnet    string        `env:"TRUSTED_SUBNET" json:"trusted_subnet"`
}

const (
	defaultServerBaseURL   = "localhost:8080"
	defaultPollInterval    = 2
	defaultReportInterval  = 10
	defaultStoreInterval   = 20
	defaultFileStoragePath = ""
	defaultRestore         = false
	defaultDatabaseDSN     = ""
	defaultSecretKey       = ""
	defaultRateLimit       = 2
	defaultRequestTimeout  = 5 * time.Second
	defaultShutdownTimeout = 10 * time.Second
)

// ========== Agent ==========

// NewAgentConfig creates an agent configuration from CLI arguments and environment variables.
func NewAgentConfig(args []string) (*AgentConfig, error) {
	fs := flag.NewFlagSet("agent", flag.ContinueOnError)

	var configPath string
	fs.StringVar(&configPath, "c", "", "path to JSON config file")
	fs.StringVar(&configPath, "config", "", "path to JSON config file")

	addr := fs.String("a", defaultServerBaseURL, "address and port")
	poll := fs.Int("p", defaultPollInterval, "poll interval")
	report := fs.Int("r", defaultReportInterval, "report interval")
	key := fs.String("k", defaultSecretKey, "secret key")
	limit := fs.Int("l", defaultRateLimit, "rate limit")
	crypto := fs.String("crypto-key", "", "path to public key")

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}

	cfg := &AgentConfig{
		ServerBaseURL:   *addr,
		PollInterval:    *poll,
		ReportInterval:  *report,
		SecretKey:       *key,
		RateLimit:       *limit,
		PublicCryptoKey: *crypto,
	}

	// 1. JSON (низший приоритет)
	if configPath == "" {
		configPath = os.Getenv("CONFIG")
	}

	if err := loadJSONFile(configPath, cfg); err != nil {
		return nil, fmt.Errorf("failed to load JSON config: %w", err)
	}

	// 2. Env (средний)
	if err := loadDotEnv(); err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse env: %w", err)
	}

	// 3. Visit гарантирует, что перезапишутся только те флаги, которые явно указали
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			cfg.ServerBaseURL = *addr
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
		}
	})

	if cfg.ServerBaseURL == "" {
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

// ========== Server ==========

func NewServerConfig(args []string) (*ServerConfig, error) {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)

	var configPath string
	fs.StringVar(&configPath, "c", "", "path to JSON config file")
	fs.StringVar(&configPath, "config", "", "path to JSON config file")

	addr := fs.String("a", defaultServerBaseURL, "address and port")
	store := fs.Int("i", defaultStoreInterval, "store interval")
	file := fs.String("f", defaultFileStoragePath, "file storage path")
	restore := fs.Bool("r", defaultRestore, "restore")
	dsn := fs.String("d", defaultDatabaseDSN, "database DSN")
	key := fs.String("k", defaultSecretKey, "secret key")
	reqTimeout := fs.Duration("rt", defaultRequestTimeout, "request timeout")
	shutTimeout := fs.Duration("s", defaultShutdownTimeout, "shutdown timeout")
	auditFile := fs.String("audit-file", "", "audit file")
	auditURL := fs.String("audit-url", "", "audit url")
	crypto := fs.String("crypto-key", "", "path to private key")
	trustedSubnet := fs.String("t", "", "network subnet in CIDR notation (e.g. 192.168.1.0/24)")

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}

	cfg := &ServerConfig{
		ServerBaseURL:    *addr,
		StoreInterval:    *store,
		FileStoragePath:  *file,
		Restore:          *restore,
		DatabaseDSN:      *dsn,
		SecretKey:        *key,
		RequestTimeout:   *reqTimeout,
		ShutdownTimeout:  *shutTimeout,
		AuditFile:        *auditFile,
		AuditURL:         *auditURL,
		PrivateCryptoKey: *crypto,
		TrustedSubnet:    *trustedSubnet,
	}

	// 1. JSON
	if configPath == "" {
		configPath = os.Getenv("CONFIG")
	}

	if err := loadJSONFile(configPath, cfg); err != nil {
		return nil, fmt.Errorf("failed to load JSON config: %w", err)
	}

	// 2. Env
	if err := loadDotEnv(); err != nil {
		return nil, fmt.Errorf("failed to load .env file: %w", err)
	}

	if err := env.Parse(cfg); err != nil {
		return nil, fmt.Errorf("failed to parse env: %w", err)
	}

	// 3. Visit гарантирует, что перезапишутся только те флаги, которые явно указали
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			cfg.ServerBaseURL = *addr
		case "i":
			cfg.StoreInterval = *store
		case "f":
			cfg.FileStoragePath = *file
		case "r":
			cfg.Restore = *restore
		case "d":
			cfg.DatabaseDSN = *dsn
		case "k":
			cfg.SecretKey = *key
		case "rt":
			cfg.RequestTimeout = *reqTimeout
		case "s":
			cfg.ShutdownTimeout = *shutTimeout
		case "audit-file":
			cfg.AuditFile = *auditFile
		case "audit-url":
			cfg.AuditURL = *auditURL
		case "crypto-key":
			cfg.PrivateCryptoKey = *crypto
		}
	})

	if cfg.ServerBaseURL == "" {
		return nil, errors.New("server URL can not be empty")
	}
	if cfg.StoreInterval < 0 {
		return nil, errors.New("store interval must be non-negative")
	}

	return cfg, nil
}

func loadJSONFile(path string, v interface{}) error {
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("could not read config file %s: %w", path, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("could not parse config file %s: %w", path, err)
	}
	return nil
}

func loadDotEnv() error {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("could not load .env file: %w", err)
	}
	return nil
}
