package main

import (
	"log"
	"metrics-collector/internal/agent"
	"metrics-collector/internal/config"
	"metrics-collector/internal/logger"
	"os"
)

func main() {
	logger, err := logger.NewSugarZapLogger()
	if err != nil {
		log.Fatalf("Failed to create agent logger: %v", err)
	}
	defer logger.Sync()

	logger.Infoln("Starting agent...")

	cfg, err := config.NewAgentConfig(os.Args[1:])
	if err != nil {
		logger.Fatalf("failed to create agent config: %v", err)
	}

	if cfg == nil {
		logger.Fatalln("Config is nil")
	}

	agent.Run(cfg, logger)

}
