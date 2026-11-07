package opentelemetry

// OpenTelemetryIntegration manages OpenTelemetry collector and instrumentation settings.
type OpenTelemetryIntegration struct {
	CollectorGRPCEndpoint string
	CollectorHTTPEndpoint string
}

func NewOpenTelemetryIntegration() *OpenTelemetryIntegration {
	return &OpenTelemetryIntegration{
		CollectorGRPCEndpoint: "otel-collector.monitoring.svc.cluster.local:4317",
		CollectorHTTPEndpoint: "otel-collector.monitoring.svc.cluster.local:4318",
	}
}
