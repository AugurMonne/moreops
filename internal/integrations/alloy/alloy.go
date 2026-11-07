package alloy

// AlloyIntegration manages Grafana Alloy telemetry collector configs.
type AlloyIntegration struct {
	ConfigPath string
}

func NewAlloyIntegration() *AlloyIntegration {
	return &AlloyIntegration{ConfigPath: "/etc/alloy/config.alloy"}
}
