package config

type DeviceRegistry struct {
	Devices map[uint64]DeviceConfig `yaml:"devices"` // NodeID → config
}

type DeviceConfig struct {
	Topic       string                 `yaml:"topic"`       // "study/presence"
	Sensitivity string                 `yaml:"sensitivity"` // Device-specific
	DebounceMs  int                    `yaml:"debounce_ms"`
	Extra       map[string]interface{} `yaml:"extra"` // Future extensibility
}
