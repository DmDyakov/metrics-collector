package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/caarlos0/env/v11"
)

type ServerConfig struct {
	HTTPAddress      string        `env:"ADDRESS" json:"address"`
	GRPCAddress      string        `env:"GRPC_ADDRESS" json:"grpc_address"`
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
	defaultStoreInterval   = 20
	defaultFileStoragePath = ""
	defaultRestore         = false
	defaultDatabaseDSN     = ""
	defaultRequestTimeout  = 5 * time.Second
	defaultShutdownTimeout = 10 * time.Second
)

func NewServerConfig(args []string) (*ServerConfig, error) {
	fs := flag.NewFlagSet("server", flag.ContinueOnError)

	var configPath string
	fs.StringVar(&configPath, "c", "", "path to JSON config file")
	fs.StringVar(&configPath, "config", "", "path to JSON config file")

	addr := fs.String("a", "localhost:8080", "address and port")
	store := fs.Int("i", defaultStoreInterval, "store interval")
	file := fs.String("f", defaultFileStoragePath, "file storage path")
	restore := fs.Bool("r", defaultRestore, "restore")
	dsn := fs.String("d", defaultDatabaseDSN, "database DSN")
	key := fs.String("k", "", "secret key")
	reqTimeout := fs.Duration("rt", defaultRequestTimeout, "request timeout")
	shutTimeout := fs.Duration("s", defaultShutdownTimeout, "shutdown timeout")
	auditFile := fs.String("audit-file", "", "audit file")
	auditURL := fs.String("audit-url", "", "audit url")
	crypto := fs.String("crypto-key", "", "path to private key")
	trustedSubnet := fs.String("t", "", "network subnet in CIDR notation")
	grpcAddr := fs.String("grpc-addr", ":50051", "gRPC server address")

	if err := fs.Parse(args); err != nil {
		return nil, fmt.Errorf("failed to parse flags: %w", err)
	}

	cfg := &ServerConfig{
		HTTPAddress:      *addr,
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
		GRPCAddress:      *grpcAddr,
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
		case "grpc-addr":
			cfg.GRPCAddress = *grpcAddr
		}
	})

	if cfg.HTTPAddress == "" {
		return nil, errors.New("server URL can not be empty")
	}
	if cfg.StoreInterval < 0 {
		return nil, errors.New("store interval must be non-negative")
	}

	return cfg, nil
}
