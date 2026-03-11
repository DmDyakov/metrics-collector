package main

import (
	"log"
	"metrics-collector/internal/agent"
	"metrics-collector/internal/config"
)

func main() {
	log.Println("Starting agent...")
	cfg, err := config.NewAgentConfig()
	if err != nil {
		log.Fatalf("Failed to create agent config: %v", err)
	}

	if cfg == nil {
		log.Fatal("Config is nil")
	}

	agent.Run(cfg)

}
