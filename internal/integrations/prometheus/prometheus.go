package prometheus

import "fmt"

// Client manages Prometheus interactions and configurations.
type Client struct {
	BaseURL string
}

// NewClient returns a new Prometheus client.
func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = "http://prometheus-operated.monitoring.svc:9090"
	}
	return &Client{BaseURL: baseURL}
}

// QueryInstant formats an instant PromQL query URL.
func (c *Client) QueryInstant(query string) string {
	return fmt.Sprintf("%s/api/v1/query?query=%s", c.BaseURL, query)
}
