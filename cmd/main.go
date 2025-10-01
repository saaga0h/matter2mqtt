package main

import (
	"flag"
	"log"
	"matter2mqtt/internal/bridge"
	"matter2mqtt/internal/config"
)

func main() {
	configPath := flag.String("config", "config.yaml", "Path to config file")
	devicesPath := flag.String("devices", "devices.yaml", "Path to devices file")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Load device registry
	registry, err := config.LoadDeviceRegistry(*devicesPath)
	if err != nil {
		log.Fatalf("Failed to load devices: %v", err)
	}

	// Create and start bridge
	b, err := bridge.NewBridge(cfg, registry)
	if err != nil {
		log.Fatalf("Failed to create bridge: %v", err)
	}

	if err := b.Start(); err != nil {
		log.Fatalf("Failed to start bridge: %v", err)
	}

	log.Println("matter2mqtt started successfully")

	// Block forever
	select {}
}
