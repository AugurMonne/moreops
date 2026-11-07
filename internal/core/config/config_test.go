package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := NewDefaultConfig("test-app")
	if cfg.Project.Name != "test-app" {
		t.Fatalf("expected project name test-app, got %s", cfg.Project.Name)
	}
	if _, ok := cfg.GetEnvironment("dev"); !ok {
		t.Fatalf("expected dev environment to exist")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "moreops.yaml")

	cfg := NewDefaultConfig("my-service")
	cfg.Project.Version = "1.2.3"
	cfg.Modules.Monitoring = &MonitoringModuleConfig{
		Enabled:    true,
		Prometheus: true,
		Grafana:    true,
	}

	if err := SaveConfig(configPath, cfg); err != nil {
		t.Fatalf("failed to save config: %v", err)
	}

	loaded, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if loaded.Project.Name != "my-service" {
		t.Errorf("expected project name my-service, got %s", loaded.Project.Name)
	}
	if loaded.Project.Version != "1.2.3" {
		t.Errorf("expected project version 1.2.3, got %s", loaded.Project.Version)
	}
	if loaded.Modules.Monitoring == nil || !loaded.Modules.Monitoring.Enabled {
		t.Errorf("expected monitoring to be enabled")
	}
	if !loaded.Modules.Monitoring.Prometheus || !loaded.Modules.Monitoring.Grafana {
		t.Errorf("expected prometheus and grafana to be true")
	}
}

func TestLoadInvalidConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "invalid.yaml")

	if err := os.WriteFile(configPath, []byte("version: 1\nproject:\n  description: no name"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadConfig(configPath)
	if err == nil {
		t.Fatal("expected error on missing project.name, got nil")
	}
}
