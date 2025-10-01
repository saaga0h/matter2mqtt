package config

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
	BaseTopic string `yaml:"base_topic"` // e.g., "matter"
}

type MatterConfig struct {
	FabricPath    string `yaml:"fabric_path"`    // ~/.matter_sdk/
	ThreadNetwork string `yaml:"thread_network"` // If needed
}

type BridgeConfig struct {
	LogLevel string `yaml:"log_level"`
	Port     int    `yaml:"port"` // For health/status endpoint
}
