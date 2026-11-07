package e2e_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"moreops/internal/core/config"
	"moreops/internal/core/engine"
	"moreops/internal/validator"

	// Register all modules via blank imports
	_ "moreops/internal/modules/cicd"
	_ "moreops/internal/modules/docker"
	_ "moreops/internal/modules/kubernetes"
	_ "moreops/internal/modules/monitoring"
	_ "moreops/internal/modules/security"
	_ "moreops/internal/modules/storage"
)

func TestFullWorkflowE2E(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get wd: %v", err)
	}

	tmpDir, err := os.MkdirTemp("", "moreops-e2e-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer func() {
		_ = os.Chdir(origDir)
		_ = os.RemoveAll(tmpDir)
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to chdir: %v", err)
	}

	cfg := config.NewDefaultConfig("e2e-service")
	cfg.Project.Description = "End-to-end integration test application"
	cfg.Project.Language = "go"

	// Configure modules
	cfg.Modules.Containers = &config.ContainersModuleConfig{
		Enabled: true,
		Engine:  "docker",
		Compose: true,
	}

	cfg.Modules.Kubernetes = &config.KubernetesModuleConfig{
		Enabled:  true,
		Tool:     "helm",
		Port:     8080,
		Replicas: 3,
		Ingress:  true,
	}

	cfg.Modules.Monitoring = &config.MonitoringModuleConfig{
		Enabled:        true,
		Prometheus:     true,
		Grafana:        true,
		MetricsPath:    "/metrics",
		ScrapeInterval: "15s",
		Dashboards:     true,
		AlertRules:     true,
	}

	cfg.Modules.Security = &config.SecurityModuleConfig{
		Enabled: true,
		Vault: &config.VaultConfig{
			Enabled:    true,
			Mode:       "in-cluster-dev",
			Address:    "http://vault.vault.svc.cluster.local:8200",
			AuthMethod: "kubernetes",
		},
		ExternalSecrets: &config.ExternalSecretsConfig{
			Enabled:         true,
			RefreshInterval: "1h",
		},
	}

	cfg.Modules.Storage = &config.StorageModuleConfig{
		Enabled:  true,
		Provider: "yandex",
		Bucket:   "e2e-backups-bucket",
		Velero: &config.VeleroConfig{
			Enabled:  true,
			Schedule: "0 3 * * *",
		},
	}

	cfg.Modules.CICD = &config.CICDModuleConfig{
		Enabled:       true,
		Provider:      "jenkins",
		RunnerType:    "cloud",
		GitOps: &config.GitOpsConfig{
			Enabled:  true,
			Tool:     "argocd",
			AutoSync: true,
		},
	}

	// 1. Engine Plan
	var outBuf bytes.Buffer
	eng := engine.NewEngine(nil, &outBuf)

	ctx := context.Background()
	plan, err := eng.Plan(ctx, cfg, "dev")
	if err != nil {
		t.Fatalf("Engine Plan failed: %v", err)
	}

	if plan.ToCreate == 0 {
		t.Errorf("Expected files to be planned for creation, got %d", plan.ToCreate)
	}

	// 2. Engine Apply
	if err := eng.Apply(ctx, plan, true); err != nil {
		t.Fatalf("Engine Apply failed: %v", err)
	}

	for _, fc := range plan.Changes {
		t.Logf("Planned file: %s (action: %s)", fc.Path, fc.Action)
	}

	// 3. Verify generated files exist
	expectedFiles := []string{
		"Dockerfile",
		".dockerignore",
		"docker-compose.yml",
		filepath.Join("charts", "e2e-service", "Chart.yaml"),
		filepath.Join("charts", "e2e-service", "values.yaml"),
		filepath.Join("charts", "e2e-service", "templates", "deployment.yaml"),
		filepath.Join("charts", "e2e-service", "templates", "service.yaml"),
		filepath.Join("charts", "e2e-service", "templates", "ingress.yaml"),
		filepath.Join("charts", "e2e-service", "templates", "monitoring", "servicemonitor.yaml"),
		filepath.Join("charts", "e2e-service", "templates", "monitoring", "prometheusrule.yaml"),
		filepath.Join("charts", "e2e-service", "templates", "monitoring", "grafana-dashboard-deep.yaml"),
		filepath.Join("charts", "e2e-service", "templates", "vault-dev.yaml"),
		filepath.Join("charts", "e2e-service", "templates", "externalsecret.yaml"),
		filepath.Join("charts", "e2e-service", "templates", "storage", "velero-backup.yaml"),
		filepath.Join("gitops", "argocd-application.yaml"),
		"Jenkinsfile",
	}

	for _, relPath := range expectedFiles {
		fullPath := filepath.Join(tmpDir, relPath)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("Expected generated file not found: %s", relPath)
		}
	}

	// 4. Validate generated manifests using validator
	val := validator.NewValidator()
	issues, err := val.ValidateDirectory(tmpDir)
	if err != nil {
		t.Fatalf("Validation error: %v", err)
	}

	for _, issue := range issues {
		if issue.Severity == validator.SeverityError {
			t.Errorf("Validation error: [%s] %s (%s)", issue.RuleID, issue.Message, issue.File)
		}
	}
}
