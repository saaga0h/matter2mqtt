// internal/config/device_registry.go
package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type DeviceRegistry struct {
	Devices map[uint64]DeviceConfig `yaml:"devices"`
}

type DeviceConfig struct {
	Topic       string                 `yaml:"topic"`
	Sensitivity string                 `yaml:"sensitivity,omitempty"`
	DebounceMs  int                    `yaml:"debounce_ms,omitempty"`
	Extra       map[string]interface{} `yaml:"extra,omitempty"`
}

func LoadDeviceRegistry(path string) (*DeviceRegistry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var registry DeviceRegistry
	if err := yaml.Unmarshal(data, &registry); err != nil {
		return nil, err
	}

	return &registry, nil
}
