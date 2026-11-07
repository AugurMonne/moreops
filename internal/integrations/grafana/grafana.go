package grafana

// DashboardMetadata contains Grafana dashboard attributes.
type DashboardMetadata struct {
	Title string
	Tags  []string
	UID   string
}

// FormatDashboardConfigMapLabel returns standard label for Grafana sidecar discovery.
func FormatDashboardConfigMapLabel() map[string]string {
	return map[string]string{
		"grafana_dashboard": "1",
	}
}
