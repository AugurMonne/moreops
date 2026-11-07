package cli

import (
	"path/filepath"
	"testing"

	"moreops/internal/core/config"
)

func TestEnvCommands(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "moreops.yaml")

	cfg := config.NewDefaultConfig("env-app")
	if err := config.SaveConfig(configPath, cfg); err != nil {
		t.Fatal(err)
	}

	// Test adding a new environment via add logic
	cfg.Environments["staging"] = config.EnvironmentConfig{
		Cluster:   "staging-cluster",
		Namespace: "staging-apps",
		Cloud:     "aws",
	}
	if err := config.SaveConfig(configPath, cfg); err != nil {
		t.Fatal(err)
	}

	loaded, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}

	if _, exists := loaded.Environments["staging"]; !exists {
		t.Errorf("expected staging environment to exist")
	}
}

func TestBumpSemVer(t *testing.T) {
	tests := []struct {
		current  string
		bumpType string
		expected string
	}{
		{"0.1.0", "patch", "0.1.1"},
		{"0.1.1", "minor", "0.2.0"},
		{"0.2.0", "major", "1.0.0"},
		{"v1.2.3", "patch", "1.2.4"},
	}

	for _, tt := range tests {
		got, err := bumpSemVer(tt.current, tt.bumpType)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != tt.expected {
			t.Errorf("bumpSemVer(%q, %q) = %s; want %s", tt.current, tt.bumpType, got, tt.expected)
		}
	}
}
