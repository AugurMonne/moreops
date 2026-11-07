package tempo

// TempoIntegration manages Grafana Tempo distributed tracing configurations.
type TempoIntegration struct {
	Endpoint string
}

func NewTempoIntegration() *TempoIntegration {
	return &TempoIntegration{
		Endpoint: "tempo.monitoring.svc.cluster.local:4317",
	}
}
