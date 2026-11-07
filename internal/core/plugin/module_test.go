package plugin

import (
	"context"
	"path/filepath"
	"testing"

	"moreops/internal/core/config"
	"moreops/internal/core/diff"
)

type mockModule struct {
	name string
}

func (m *mockModule) Name() string        { return m.name }
func (m *mockModule) Description() string { return "Mock module for tests" }
func (m *mockModule) Plan(ctx context.Context, cfg *config.Config, envName string) (*PlanResult, error) {
	return &PlanResult{ModuleName: m.name}, nil
}
func (m *mockModule) Apply(ctx context.Context, plan *PlanResult) error {
	return ExecutePlan(plan)
}

func TestPluginRegistry(t *testing.T) {
	ClearRegistry()
	mod := &mockModule{name: "test-module"}
	Register(mod)

	found, ok := Get("test-module")
	if !ok || found.Name() != "test-module" {
		t.Fatalf("expected to find test-module")
	}

	all := All()
	if len(all) != 1 {
		t.Fatalf("expected 1 module, got %d", len(all))
	}
}

func TestExecutePlan(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "k8s", "deployment.yaml")

	plan := &PlanResult{
		ModuleName: "k8s",
		FileChanges: []diff.FileChange{
			{
				Path:    filePath,
				NewData: "apiVersion: apps/v1\nkind: Deployment\n",
				Action:  diff.ActionCreate,
			},
		},
	}

	if err := ExecutePlan(plan); err != nil {
		t.Fatalf("execute plan failed: %v", err)
	}

	change, err := diff.ComputeFileChange(filePath, "apiVersion: apps/v1\nkind: Deployment\n")
	if err != nil {
		t.Fatal(err)
	}
	if change.Action != diff.ActionNone {
		t.Fatalf("expected file to match, got action %s", change.Action)
	}
}
