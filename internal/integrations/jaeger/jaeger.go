package jaeger

// JaegerIntegration manages Jaeger tracing configurations.
type JaegerIntegration struct {
	CollectorEndpoint string
}

func NewJaegerIntegration() *JaegerIntegration {
	return &JaegerIntegration{
		CollectorEndpoint: "http://jaeger-collector.monitoring.svc.cluster.local:14268/api/traces",
	}
}
