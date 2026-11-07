package environment

import "moreops/internal/core/config"

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Resolve(cfg *config.Config, envName string) config.EnvironmentConfig {
	if cfg.Environments != nil {
		if env, ok := cfg.Environments[envName]; ok {
			return env
		}
	}
	return config.EnvironmentConfig{
		Namespace: "default",
	}
}
