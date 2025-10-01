package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	MQTT   MQTTConfig   `yaml:"mqtt"`
	Matter MatterConfig `yaml:"matter"`
	Bridge BridgeConfig `yaml:"bridge"`
}

type MQTTConfig struct {
	Server    string `yaml:"server"`
	Port      int    `yaml:"port"`
	User      string `yaml:"user"`
	Password  string `yaml:"password"`
	BaseTopic string `yaml:"base_topic"`
}

type MatterConfig struct {
	StoragePath string `yaml:"storage_path"`
}

type BridgeConfig struct {
	LogLevel string `yaml:"log_level"`
	Port     int    `yaml:"port"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
