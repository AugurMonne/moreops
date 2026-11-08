package docker

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"moreops/internal/core/config"
)

func TestDockerPlanDefault(t *testing.T) {
	m := &Module{}
	cfg := &config.Config{
		Project: config.ProjectMeta{
			Name:     "testapp",
			Language: "go",
		},
		Modules: config.ModulesConfig{
			Containers: &config.ContainersModuleConfig{
				Enabled: true,
				Engine:  "docker",
				Compose: true,
			},
		},
	}

	plan, err := m.Plan(context.Background(), cfg, "dev")
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}

	if len(plan.FileChanges) < 3 {
		t.Fatalf("expected at least 3 file changes, got %d", len(plan.FileChanges))
	}

	// Verify Dockerfile has non-root and trimpath
	var foundDockerfile bool
	for _, fc := range plan.FileChanges {
		if fc.Path == "Dockerfile" {
			foundDockerfile = true
			if !strings.Contains(fc.NewData, "USER 65532:65532") {
				t.Error("expected non-root user in Dockerfile")
			}
		}
	}
	if !foundDockerfile {
		t.Error("Dockerfile change not found")
	}
}

func TestDockerPlanCustomTemplate(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	defer os.Chdir(origDir)

	// Create custom template
	customDir := filepath.Join(".moreops", "templates", "docker")
	os.MkdirAll(customDir, 0755)
	customDockerfile := "# CUSTOM DOCKERFILE FOR OPEN SOURCE\nFROM alpine:latest\nCMD [\"echo\", \"custom\"]\n"
	os.WriteFile(filepath.Join(customDir, "go"), []byte(customDockerfile), 0644)

	m := &Module{}
	cfg := &config.Config{
		Project: config.ProjectMeta{
			Name:     "customapp",
			Language: "go",
		},
		Modules: config.ModulesConfig{
			Containers: &config.ContainersModuleConfig{
				Enabled: true,
				Engine:  "docker",
			},
		},
	}

	plan, err := m.Plan(context.Background(), cfg, "dev")
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}

	for _, fc := range plan.FileChanges {
		if fc.Path == "Dockerfile" {
			if !strings.Contains(fc.NewData, "# CUSTOM DOCKERFILE FOR OPEN SOURCE") {
				t.Errorf("custom template was not applied: %s", fc.NewData)
			}
		}
	}
}

func TestDockerPlanWithDatabaseAndRabbitMQ(t *testing.T) {
	m := &Module{}
	cfg := &config.Config{
		Project: config.ProjectMeta{
			Name:     "fullstack-app",
			Language: "go",
		},
		Modules: config.ModulesConfig{
			Containers: &config.ContainersModuleConfig{
				Enabled: true,
				Engine:  "docker",
				Compose: true,
			},
			Database: &config.DatabaseModuleConfig{
				Enabled: true,
				Engine:  "postgres",
			},
			Messaging: &config.MessagingModuleConfig{
				Enabled: true,
				Broker:  "rabbitmq",
			},
		},
	}

	plan, err := m.Plan(context.Background(), cfg, "dev")
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}

	var composeContent string
	for _, fc := range plan.FileChanges {
		if fc.Path == "docker-compose.yml" {
			composeContent = fc.NewData
		}
	}

	if composeContent == "" {
		t.Fatal("docker-compose.yml not found in plan")
	}

	if !strings.Contains(composeContent, "image: postgres:16-alpine") {
		t.Error("expected postgres service in docker-compose.yml")
	}
	if !strings.Contains(composeContent, "image: rabbitmq:3.13-management-alpine") {
		t.Error("expected rabbitmq service in docker-compose.yml")
	}
	if !strings.Contains(composeContent, "pgdata:") {
		t.Error("expected pgdata volume in docker-compose.yml")
	}
	if !strings.Contains(composeContent, "condition: service_healthy") {
		t.Error("expected service_healthy check in docker-compose.yml")
	}
}
