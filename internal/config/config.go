// Package config предоставляет функции для загрузки и парсинга конфигурации
// из json файла, переменных окружения и флагов командной строки.
package config

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
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
	ConfigFile      string `json:"-"`
}

type ServerConfig struct {
	ServerBaseURL    string        `env:"ADDRESS" json:"address"`
	StoreInterval    int           `env:"STORE_INTERVAL" json:"store_interval"`
	FileStoragePath  string        `env:"FILE_STORAGE_PATH" json:"file_storage_path"`
	Restore          bool          `env:"RESTORE" json:"restore"`
	DatabaseDSN      string        `env:"DATABASE_DSN" json:"database_dsn"`
	SecretKey        string        `env:"KEY" json:"key"`
	PrivateCryptoKey string        `env:"CRYPTO_KEY" json:"crypto_key"`
	RequestTimeout   time.Duration `env:"REQ_TIMEOUT" json:"req_timeout"`
	ShutdownTimeout  time.Duration `env:"SHUTDOWN_TIMEOUT" json:"shutdown_timeout"`
	AuditFile        string        `env:"AUDIT_FILE" json:"audit_file"`
	AuditURL         string        `env:"AUDIT_URL" json:"audit_url"`
	PprofAddr        string        `env:"PPROF_ADDR" json:"pprof_addr"`
	ConfigFile       string        `json:"-"`
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

// NewAgentConfig creates an agent configuration from CLI arguments, environment variables, and JSON config file.
// Priority order (highest to lowest): flags -> env -> json -> defaults
func NewAgentConfig(args []string) (*AgentConfig, error) {
	loadDotEnv()

	// Start with defaults
	cfg := &AgentConfig{
		ServerBaseURL:  defaultServerBaseURL,
		PollInterval:   defaultPollInterval,
		ReportInterval: defaultReportInterval,
		SecretKey:      defaultSecretKey,
		RateLimit:      defaultRateLimit,
	}

	fs := flag.NewFlagSet("agent", flag.ContinueOnError)
	fs.StringVar(&cfg.ConfigFile, "c", "", "path to JSON config file")
	fs.StringVar(&cfg.ServerBaseURL, "a", cfg.ServerBaseURL, "address and port to run server")
	fs.IntVar(&cfg.PollInterval, "p", cfg.PollInterval, "poll interval")
	fs.IntVar(&cfg.ReportInterval, "r", cfg.ReportInterval, "report interval")
	fs.StringVar(&cfg.SecretKey, "k", cfg.SecretKey, "secret key")
	fs.IntVar(&cfg.RateLimit, "l", cfg.RateLimit, "rate limit")
	fs.StringVar(&cfg.PublicCryptoKey, "crypto-key", "", "path to public key file for encryption")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// Load JSON config if specified (lowest priority)
	if cfg.ConfigFile != "" {
		if err := loadJSONConfig(cfg.ConfigFile, cfg); err != nil {
			return nil, fmt.Errorf("failed to load JSON config: %w", err)
		}
	}

	// Apply environment variables (medium priority)
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	// Re-apply flags to ensure they have highest priority
	// We need to parse flags again or track which flags were explicitly set
	fs = flag.NewFlagSet("agent", flag.ContinueOnError)
	fs.StringVar(&cfg.ConfigFile, "c", cfg.ConfigFile, "path to JSON config file")
	flagAddr := fs.String("a", cfg.ServerBaseURL, "")
	flagPoll := fs.Int("p", cfg.PollInterval, "")
	flagReport := fs.Int("r", cfg.ReportInterval, "")
	flagKey := fs.String("k", cfg.SecretKey, "")
	flagLimit := fs.Int("l", cfg.RateLimit, "")
	flagCrypto := fs.String("crypto-key", cfg.PublicCryptoKey, "")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// Check which flags were explicitly set and apply them
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			cfg.ServerBaseURL = *flagAddr
		case "p":
			cfg.PollInterval = *flagPoll
		case "r":
			cfg.ReportInterval = *flagReport
		case "k":
			cfg.SecretKey = *flagKey
		case "l":
			cfg.RateLimit = *flagLimit
		case "crypto-key":
			cfg.PublicCryptoKey = *flagCrypto
		}
	})

	// Validation
	if cfg.ServerBaseURL == "" {
		return nil, errors.New("server URL can not be empty")
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
		return nil, fmt.Errorf("report interval (%d) must be greater than or equal to poll interval (%d)", cfg.ReportInterval, cfg.PollInterval)
	}

	return cfg, nil
}

// NewServerConfig creates a server configuration from CLI arguments, environment variables, and JSON config file.
// Priority order (highest to lowest): flags -> env -> json -> defaults
func NewServerConfig(args []string) (*ServerConfig, error) {
	loadDotEnv()

	// Start with defaults
	cfg := &ServerConfig{
		ServerBaseURL:   defaultServerBaseURL,
		StoreInterval:   defaultStoreInterval,
		FileStoragePath: defaultFileStoragePath,
		Restore:         defaultRestore,
		DatabaseDSN:     defaultDatabaseDSN,
		SecretKey:       defaultSecretKey,
	}

	fs := flag.NewFlagSet("server", flag.ContinueOnError)
	fs.StringVar(&cfg.ConfigFile, "c", "", "path to JSON config file")
	fs.StringVar(&cfg.ServerBaseURL, "a", cfg.ServerBaseURL, "address and port to run server")
	fs.IntVar(&cfg.StoreInterval, "i", cfg.StoreInterval, "store interval")
	fs.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "file storage path")
	fs.BoolVar(&cfg.Restore, "r", cfg.Restore, "restore")
	fs.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "database DSN")
	fs.StringVar(&cfg.SecretKey, "k", cfg.SecretKey, "secret key")
	fs.DurationVar(&cfg.RequestTimeout, "t", cfg.RequestTimeout, "request timeout")
	fs.DurationVar(&cfg.ShutdownTimeout, "s", cfg.ShutdownTimeout, "shutdown timeout")
	fs.StringVar(&cfg.AuditFile, "audit-file", cfg.AuditFile, "audit file for logs")
	fs.StringVar(&cfg.AuditURL, "audit-url", cfg.AuditURL, "audit url for logs")
	fs.StringVar(&cfg.PrivateCryptoKey, "crypto-key", "", "path to private key file for decryption")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// Load JSON config if specified (lowest priority)
	if cfg.ConfigFile != "" {
		if err := loadJSONConfig(cfg.ConfigFile, cfg); err != nil {
			return nil, fmt.Errorf("failed to load JSON config: %w", err)
		}
	}

	// Apply environment variables (medium priority)
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	// Re-apply flags to ensure they have highest priority
	fs = flag.NewFlagSet("server", flag.ContinueOnError)
	fs.StringVar(&cfg.ConfigFile, "c", cfg.ConfigFile, "path to JSON config file")
	flagAddr := fs.String("a", cfg.ServerBaseURL, "")
	flagStore := fs.Int("i", cfg.StoreInterval, "")
	flagFile := fs.String("f", cfg.FileStoragePath, "")
	flagRestore := fs.Bool("r", cfg.Restore, "")
	flagDSN := fs.String("d", cfg.DatabaseDSN, "")
	flagKey := fs.String("k", cfg.SecretKey, "")
	flagTimeout := fs.Duration("t", cfg.RequestTimeout, "")
	flagShutdown := fs.Duration("s", cfg.ShutdownTimeout, "")
	flagAuditFile := fs.String("audit-file", cfg.AuditFile, "")
	flagAuditURL := fs.String("audit-url", cfg.AuditURL, "")
	flagCrypto := fs.String("crypto-key", cfg.PrivateCryptoKey, "")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	// Check which flags were explicitly set and apply them
	fs.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			cfg.ServerBaseURL = *flagAddr
		case "i":
			cfg.StoreInterval = *flagStore
		case "f":
			cfg.FileStoragePath = *flagFile
		case "r":
			cfg.Restore = *flagRestore
		case "d":
			cfg.DatabaseDSN = *flagDSN
		case "k":
			cfg.SecretKey = *flagKey
		case "t":
			cfg.RequestTimeout = *flagTimeout
		case "s":
			cfg.ShutdownTimeout = *flagShutdown
		case "audit-file":
			cfg.AuditFile = *flagAuditFile
		case "audit-url":
			cfg.AuditURL = *flagAuditURL
		case "crypto-key":
			cfg.PrivateCryptoKey = *flagCrypto
		}
	})

	// Validation
	if cfg.ServerBaseURL == "" {
		return nil, errors.New("server URL can not be empty")
	}

	if cfg.StoreInterval < 0 {
		return nil, errors.New("store interval must be non-negative")
	}

	return cfg, nil
}

// loadDotEnv loads variables from a .env file if it exists.
func loadDotEnv() {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("Warning: could not load .env file: %v", err)
	}
}

// loadJSONConfig loads configuration from a JSON file.
func loadJSONConfig(path string, cfg interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(data, cfg); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	return nil
}
