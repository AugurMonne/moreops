package flux

import "fmt"

// KustomizationOptions holds parameters for Flux CD Kustomization.
type KustomizationOptions struct {
	Name      string
	Namespace string
	Path      string
	Interval  string
}

// GenerateKustomization returns Flux CD Kustomization YAML.
func GenerateKustomization(opts KustomizationOptions) string {
	interval := opts.Interval
	if interval == "" {
		interval = "5m"
	}
	return fmt.Sprintf(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: %s
  namespace: flux-system
spec:
  interval: %s
  path: "%s"
  prune: true
  targetNamespace: %s
  sourceRef:
    kind: GitRepository
    name: %s-repo
`, opts.Name, interval, opts.Path, opts.Namespace, opts.Name)
}
