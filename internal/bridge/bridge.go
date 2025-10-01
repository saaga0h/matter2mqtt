package bridge

import (
	"fmt"
	"log" // Add this import

	"matter2mqtt/internal/config"
	"matter2mqtt/internal/matter"
	"matter2mqtt/internal/mqtt"
)

type Bridge struct {
	config         *config.Config
	deviceRegistry *config.DeviceRegistry
	matterClient   *matter.Client
	mqttClient     *mqtt.Client
	devices        map[uint64]*Device
}

func NewBridge(cfg *config.Config, registry *config.DeviceRegistry) (*Bridge, error) {
	matterClient, err := matter.NewClient(cfg.Matter.StoragePath)
	if err != nil {
		return nil, err
	}

	mqttClient, err := mqtt.NewClient(cfg.MQTT)
	if err != nil {
		return nil, err
	}

	return &Bridge{
		config:         cfg,
		deviceRegistry: registry,
		matterClient:   matterClient,
		mqttClient:     mqttClient,
		devices:        make(map[uint64]*Device),
	}, nil
}

func (b *Bridge) Start() error {
	// Connect MQTT
	if err := b.mqttClient.Connect(); err != nil {
		return err
	}

	// Publish bridge status
	b.publishBridgeState("online")

	// Initialize all configured devices
	for nodeID, devCfg := range b.deviceRegistry.Devices {
		if err := b.initializeDevice(nodeID, devCfg); err != nil {
			log.Printf("Failed to initialize device %d: %v", nodeID, err)
			continue
		}
	}

	return nil
}

func (b *Bridge) initializeDevice(nodeID uint64, cfg config.DeviceConfig) error {
	// Connect to Matter device
	session, err := b.matterClient.Connect(nodeID)
	if err != nil {
		return err
	}

	// Create device handler
	device := NewDevice(nodeID, cfg, session, b.mqttClient)

	// Query device capabilities (interview)
	if err := device.Interview(); err != nil {
		return err
	}

	// Setup subscriptions based on capabilities
	if err := device.SetupSubscriptions(); err != nil {
		return err
	}

	b.devices[nodeID] = device
	return nil
}

func (b *Bridge) publishBridgeState(state string) {
	topic := fmt.Sprintf("%s/bridge/state", b.config.MQTT.BaseTopic)
	b.mqttClient.Publish(topic, state, true) // retained
}
