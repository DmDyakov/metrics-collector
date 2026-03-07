package main

import (
	"log"
	"metrics-collector/internal/agent"
	"metrics-collector/internal/config"
)

func main() {
	log.Println("Starting agent...")
	cfg := config.NewConfig()

	agent.Run(cfg)

}
