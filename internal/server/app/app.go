// Package app собирает и запускает все серверы.
package app

import (
	"context"
	"fmt"
	"io"

	"metrics-collector/internal/config"
	"metrics-collector/internal/domain/audit"
	"metrics-collector/internal/server/repository"
	"metrics-collector/internal/server/repository/file"
	"metrics-collector/internal/server/repository/mem"
	"metrics-collector/internal/server/repository/postgres"
	"metrics-collector/internal/server/service"
	"metrics-collector/internal/server/transport/grpc"
	"metrics-collector/internal/server/transport/http"
	"metrics-collector/internal/server/transport/http/handler"
	"metrics-collector/internal/server/worker"
	"metrics-collector/pkg/encryptor"
	"metrics-collector/pkg/publisher"
	"metrics-collector/pkg/signer"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

// App управляет жизненным циклом сервера.
type App struct {
	cfg               *config.ServerConfig
	logger            *zap.Logger
	httpServer        *http.Server
	grpcServer        *grpc.Server
	persistentStorage io.Closer
	auditPub          *publisher.Publisher[audit.Event]
	backupWorker      *worker.BackupWorker
}

// New создаёт новый App.
func New(cfg *config.ServerConfig, logger *zap.Logger) (*App, error) {
	// Хранилища
	memStorage := mem.NewMemStorage()
	persistentStorage, err := initPersistentStorage(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize persistentStorage: %w", err)
	}

	repo := repository.New(memStorage, persistentStorage, cfg.StoreInterval, logger)

	// Сервисы
	metricsService := service.NewMetricsService(repo)
	healthService := service.NewHealthService(repo)

	// Publisher событий для аудита
	auditPub, err := publisher.NewPublisher[audit.Event](cfg.AuditFile, cfg.AuditURL, cfg.RequestTimeout, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create audit publisher: %w", err)
	}

	// Криптография
	signer, err := signer.New(cfg.SecretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize signer: %w", err)
	}
	enc, err := encryptor.New("", cfg.PrivateCryptoKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create encryptor: %w", err)
	}

	// HTTP сервер
	metricsHandler, err := handler.NewMetricsHandler(metricsService, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to create metrics handler: %w", err)
	}
	healthHandler := handler.NewHealthHandler(healthService, logger)

	httpServer := http.NewServer(
		cfg,
		healthHandler,
		metricsHandler,
		auditPub,
		signer,
		enc,
		logger,
	)

	// gRPC сервер
	var grpcServer *grpc.Server
	if cfg.GRPCAddress != "" {
		grpcHandler := grpc.NewHandler(metricsService, logger)
		grpcServer, err = grpc.NewServer(
			cfg.GRPCAddress,
			cfg.TrustedSubnet,
			grpcHandler,
			logger,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create gRPC server: %w", err)
		}
	}

	// Воркеры
	backupWorker := worker.NewBackupWorker(
		cfg.Restore,
		cfg.StoreInterval,
		cfg.RequestTimeout,
		memStorage,
		persistentStorage,
		logger,
	)

	return &App{
		cfg:               cfg,
		logger:            logger,
		httpServer:        httpServer,
		grpcServer:        grpcServer,
		persistentStorage: persistentStorage,
		auditPub:          auditPub,
		backupWorker:      backupWorker,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	defer a.persistentStorage.Close()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return a.httpServer.Run(ctx)
	})

	g.Go(func() error {
		return a.grpcServer.Run(ctx)
	})

	g.Go(func() error {
		return a.backupWorker.Run(ctx)
	})

	g.Go(func() error {
		return a.auditPub.Run(ctx)
	})

	return g.Wait()
}

func initPersistentStorage(cfg *config.ServerConfig, logger *zap.Logger) (repository.PersistentStorage, error) {
	switch {
	case cfg.DatabaseDSN != "":
		pg, _, err := postgres.NewPostgresStorage(cfg.DatabaseDSN, logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create postgres storage: %w", err)
		}
		logger.Info("Storage mode: postgres")
		return pg, nil

	case cfg.FileStoragePath != "":
		logger.Info("Storage mode: file")
		return file.NewFileStorage(cfg.FileStoragePath), nil

	default:
		logger.Info("Storage mode: memory only")
		return nil, nil
	}
}
