package engine

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"moreops/internal/core/config"
	"moreops/internal/core/diff"
	"moreops/internal/core/plugin"
)

type testModule struct {
	targetFile string
	content    string
}

func (m *testModule) Name() string        { return "test-mod" }
func (m *testModule) Description() string { return "test module" }
func (m *testModule) Plan(ctx context.Context, cfg *config.Config, envName string) (*plugin.PlanResult, error) {
	fc, err := diff.ComputeFileChange(m.targetFile, m.content)
	if err != nil {
		return nil, err
	}
	return &plugin.PlanResult{
		ModuleName:  m.Name(),
		FileChanges: []diff.FileChange{fc},
	}, nil
}
func (m *testModule) Apply(ctx context.Context, plan *plugin.PlanResult) error {
	return plugin.ExecutePlan(plan)
}

func TestEnginePlanAndApplyConfirm(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "k8s", "service.yaml")

	plugin.ClearRegistry()
	mod := &testModule{
		targetFile: targetPath,
		content:    "apiVersion: v1\nkind: Service\n",
	}
	plugin.Register(mod)

	in := strings.NewReader("yes\n")
	out := &bytes.Buffer{}
	eng := NewEngine(in, out)

	cfg := config.NewDefaultConfig("test")
	plan, err := eng.Plan(context.Background(), cfg, "dev")
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}

	if plan.ToCreate != 1 {
		t.Fatalf("expected 1 file to create, got %d", plan.ToCreate)
	}

	if err := eng.Apply(context.Background(), plan, false); err != nil {
		t.Fatalf("Apply failed: %v", err)
	}

	// Verify file was written
	data, err := os.ReadFile(targetPath)
	if err != nil {
		t.Fatalf("file not found after apply: %v", err)
	}
	if string(data) != "apiVersion: v1\nkind: Service\n" {
		t.Fatalf("file content mismatch: %s", string(data))
	}
}

func TestEngineApplyCancelled(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "k8s", "deployment.yaml")

	plugin.ClearRegistry()
	mod := &testModule{
		targetFile: targetPath,
		content:    "apiVersion: apps/v1\nkind: Deployment\n",
	}
	plugin.Register(mod)

	in := strings.NewReader("no\n")
	out := &bytes.Buffer{}
	eng := NewEngine(in, out)

	cfg := config.NewDefaultConfig("test")
	plan, err := eng.Plan(context.Background(), cfg, "dev")
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}

	if err := eng.Apply(context.Background(), plan, false); err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}

	// Verify file was NOT written
	if _, err := os.Stat(targetPath); !os.IsNotExist(err) {
		t.Fatalf("file should not exist when apply was cancelled")
	}
}
