package developer

type ToolingConfig struct {
	HotReload bool `yaml:"hot_reload"`
	DebugPort int  `yaml:"debug_port"`
}

func DefaultTooling() *ToolingConfig {
	return &ToolingConfig{
		HotReload: true,
		DebugPort: 40000,
	}
}
