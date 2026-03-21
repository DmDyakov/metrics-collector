package main

import (
	"log"
	"metrics-collector/internal/agent"
	"metrics-collector/internal/config"
	"os"
)

func main() {
	log.Println("Starting agent...")
	cfg, err := config.NewAgentConfig(os.Args[1:])
	if err != nil {
		log.Fatalf("Failed to create agent config: %v", err)
	}

	if cfg == nil {
		log.Fatal("Config is nil")
	}

	agent.Run(cfg)

}
