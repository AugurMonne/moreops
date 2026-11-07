package unit_test

import (
	"testing"

	"moreops/internal/core/config"
	"moreops/internal/detector"
)

func TestDefaultConfigGeneration(t *testing.T) {
	cfg := config.NewDefaultConfig("test-app")
	if cfg.Project.Name != "test-app" {
		t.Errorf("Expected project name 'test-app', got '%s'", cfg.Project.Name)
	}

	if cfg.Environments["dev"].Namespace != "default" {
		t.Errorf("Expected dev namespace 'default', got '%s'", cfg.Environments["dev"].Namespace)
	}
}

func TestDetectorBasics(t *testing.T) {
	det := detector.NewDetector()
	if det == nil {
		t.Fatal("Expected detector instance, got nil")
	}
}
