package bridge

import (
	"encoding/json"
	"fmt"
	"matter2mqtt/internal/config"
	"matter2mqtt/internal/matter"
	"matter2mqtt/internal/mqtt"
	"time"
)

type Device struct {
	nodeID        uint64
	config        config.DeviceConfig
	session       *matter.Session
	mqttClient    *mqtt.Client
	capabilities  []matter.ClusterAttribute
	subscriptions []*matter.Subscription
}

func NewDevice(nodeID uint64, cfg config.DeviceConfig, session *matter.Session, mqtt *mqtt.Client) *Device {
	return &Device{
		nodeID:     nodeID,
		config:     cfg,
		session:    session,
		mqttClient: mqtt,
	}
}

func (d *Device) Interview() error {
	// Query device for supported clusters/attributes
	// This would read descriptor cluster (0x001D) to discover capabilities
	// For now, we could have a static mapping based on device type

	// Example: assume it's a presence sensor
	d.capabilities = []matter.ClusterAttribute{
		{ClusterID: matter.ClusterOccupancySensing, AttributeID: matter.AttributeOccupancy, Name: "presence", DataType: "bool"},
		{ClusterID: matter.ClusterIlluminanceMeasurement, AttributeID: 0x0000, Name: "illuminance", DataType: "uint16"},
		{ClusterID: matter.ClusterPowerSource, AttributeID: 0x000C, Name: "battery", DataType: "uint8"},
	}

	return nil
}

func (d *Device) SetupSubscriptions() error {
	for _, cap := range d.capabilities {
		sub, err := d.session.Subscribe(cap.ClusterID, cap.AttributeID, func(value interface{}) {
			d.handleAttributeChange(cap.Name, value)
		})

		if err != nil {
			return fmt.Errorf("failed to subscribe to %s: %w", cap.Name, err)
		}

		d.subscriptions = append(d.subscriptions, sub)
	}

	return nil
}

func (d *Device) handleAttributeChange(attribute string, value interface{}) {
	// Convert to MQTT payload
	payload := d.convertToPayload(attribute, value)

	// Publish to MQTT
	topic := fmt.Sprintf("matter/%s", d.config.Topic)
	d.mqttClient.Publish(topic, payload, false)
}

func (d *Device) convertToPayload(attribute string, value interface{}) string {
	// Build JSON payload
	data := map[string]interface{}{
		attribute:   value,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"node_id":   d.nodeID,
	}

	jsonData, _ := json.Marshal(data)
	return string(jsonData)
}
