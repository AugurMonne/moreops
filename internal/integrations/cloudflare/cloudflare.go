package cloudflare

// CloudflareIntegration manages DNS, Ingress tunnels, and SSL certificates.
type CloudflareIntegration struct {
	ZoneID string
}

func NewCloudflareIntegration(zoneID string) *CloudflareIntegration {
	return &CloudflareIntegration{ZoneID: zoneID}
}
