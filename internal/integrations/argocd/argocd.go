package argocd

import "fmt"

// AppOptions holds parameters for Argo CD Application manifest.
type AppOptions struct {
	Name        string
	Namespace   string
	RepoURL     string
	Path        string
	TargetEnv   string
	Destination string
}

// FormatSyncPolicy returns automated sync policy YAML.
func FormatSyncPolicy() string {
	return `    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
      - ApplyOutOfSyncOnly=true`
}

// DefaultServer returns default in-cluster k8s API server URL.
func DefaultServer() string {
	return "https://kubernetes.default.svc"
}

// AppName returns formatted Argo CD application name.
func AppName(app, env string) string {
	return fmt.Sprintf("%s-%s", app, env)
}
