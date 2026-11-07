package loki

// LokiIntegration provides configuration generators for Grafana Loki log aggregation.
type LokiIntegration struct {
	Endpoint string
}

func NewLokiIntegration(endpoint string) *LokiIntegration {
	if endpoint == "" {
		endpoint = "http://loki-gateway.monitoring.svc.cluster.local"
	}
	return &LokiIntegration{Endpoint: endpoint}
}
