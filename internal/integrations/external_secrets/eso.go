package external_secrets

// ESOIntegration configures External Secrets Operator resources.
type ESOIntegration struct {
	Namespace string
}

func NewESOIntegration(ns string) *ESOIntegration {
	if ns == "" {
		ns = "external-secrets"
	}
	return &ESOIntegration{Namespace: ns}
}
