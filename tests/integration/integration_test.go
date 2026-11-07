package integration_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"moreops/internal/core/config"
	"moreops/internal/core/engine"

	_ "moreops/internal/modules/cicd"
	_ "moreops/internal/modules/docker"
	_ "moreops/internal/modules/kubernetes"
	_ "moreops/internal/modules/monitoring"
	_ "moreops/internal/modules/security"
	_ "moreops/internal/modules/storage"
)

func TestMultiEnvironmentPlanning(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get wd: %v", err)
	}

	tmpDir, err := os.MkdirTemp("", "moreops-multienv-*")
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

	cfg := config.NewDefaultConfig("multienv-app")
	cfg.Environments["staging"] = config.EnvironmentConfig{
		Cluster:   "k8s-staging-cluster",
		Namespace: "staging",
		Cloud:     "yandex",
	}
	cfg.Environments["prod"] = config.EnvironmentConfig{
		Cluster:   "k8s-prod-cluster",
		Namespace: "production",
		Cloud:     "yandex",
	}

	cfg.Modules.Kubernetes = &config.KubernetesModuleConfig{
		Enabled:  true,
		Tool:     "helm",
		Port:     8080,
		Replicas: 2,
		Ingress:  true,
	}

	eng := engine.NewEngine(nil, nil)
	ctx := context.Background()

	// Plan for dev
	devPlan, err := eng.Plan(ctx, cfg, "dev")
	if err != nil {
		t.Fatalf("Plan for dev failed: %v", err)
	}
	if devPlan.Environment != "dev" {
		t.Errorf("Expected environment dev, got %s", devPlan.Environment)
	}

	// Plan for prod
	prodPlan, err := eng.Plan(ctx, cfg, "prod")
	if err != nil {
		t.Fatalf("Plan for prod failed: %v", err)
	}
	if prodPlan.Environment != "prod" {
		t.Errorf("Expected environment prod, got %s", prodPlan.Environment)
	}

	// Apply dev plan
	if err := eng.Apply(ctx, devPlan, true); err != nil {
		t.Fatalf("Apply dev plan failed: %v", err)
	}

	// Ensure Chart.yaml was created
	chartFile := filepath.Join(tmpDir, "charts", "multienv-app", "Chart.yaml")
	if _, err := os.Stat(chartFile); os.IsNotExist(err) {
		t.Fatalf("Chart.yaml was not generated")
	}
}
