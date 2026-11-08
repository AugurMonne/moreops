package wizard

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestParseSelection(t *testing.T) {
	tests := []struct {
		input    string
		max      int
		expected []int
	}{
		{"1 4", 6, []int{1, 4}},
		{"1,4", 6, []int{1, 4}},
		{"1, 4", 6, []int{1, 4}},
		{"14", 6, []int{1, 4}},
		{"25", 6, []int{2, 5}},
		{"136", 6, []int{1, 3, 6}},
		{"", 6, nil},
		{"9", 6, nil}, // out of bounds
	}

	for _, tt := range tests {
		got := ParseSelection(tt.input, tt.max)
		if !reflect.DeepEqual(got, tt.expected) {
			t.Errorf("ParseSelection(%q, %d) = %v; want %v", tt.input, tt.max, got, tt.expected)
		}
	}
}

func TestWizardRunDefaults(t *testing.T) {
	// Simulating pressing Enter for all questions:
	// 1. Project name -> Enter (default: my-app)
	// 2. Version -> Enter (default: 0.1.0)
	// 3. Env -> Enter (default: dev)
	// 4. Cluster -> Enter (default: current-context)
	// 5. Namespace -> Enter (default: default)
	// 6. Selection -> Enter (default: 1 2)
	// 7. Kubernetes strategy -> Enter (default: helm)
	// 8. Replicas count -> Enter (default: 2)
	// 9. Enable HPA -> Enter (default: yes)
	// 10. Max replicas -> Enter (default: 10)
	simulatedInput := "\n\n\n\n\n\n\n\n\n\n"
	in := strings.NewReader(simulatedInput)
	out := &bytes.Buffer{}

	w := NewWizard(in, out)
	cfg, err := w.Run("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Project.Name == "" {
		t.Errorf("expected non-empty project name")
	}
	if cfg.Project.Version != "0.1.0" {
		t.Errorf("expected version 0.1.0, got %s", cfg.Project.Version)
	}
	if cfg.Modules.Containers == nil || !cfg.Modules.Containers.Enabled {
		t.Errorf("expected containers module enabled")
	}
	if cfg.Modules.Kubernetes == nil || !cfg.Modules.Kubernetes.Enabled {
		t.Errorf("expected kubernetes module enabled")
	}
	if cfg.Modules.Kubernetes.Tool != "helm" {
		t.Errorf("expected default k8s tool helm, got %s", cfg.Modules.Kubernetes.Tool)
	}
	if cfg.Modules.Monitoring != nil {
		t.Errorf("expected monitoring module to be nil")
	}
}

func TestWizardRunCustomSelection(t *testing.T) {
	// Enter project name "custom-srv", version "2.0.0", env "prod", cluster "k8s-prod", ns "apps", selection "14", vault -> enter, rbac -> enter
	simulatedInput := "custom-srv\n2.0.0\nprod\nk8s-prod\napps\n14\n\n\n"
	in := strings.NewReader(simulatedInput)
	out := &bytes.Buffer{}

	w := NewWizard(in, out)
	cfg, err := w.Run("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Project.Name != "custom-srv" {
		t.Errorf("expected custom-srv, got %s", cfg.Project.Name)
	}
	if cfg.Project.Version != "2.0.0" {
		t.Errorf("expected 2.0.0, got %s", cfg.Project.Version)
	}
	env, ok := cfg.GetEnvironment("prod")
	if !ok || env.Namespace != "apps" {
		t.Errorf("expected prod environment with namespace apps")
	}
	// Selection "14": 1 (Containers) and 4 (Security)
	if cfg.Modules.Containers == nil || !cfg.Modules.Containers.Enabled {
		t.Errorf("expected containers module to be enabled")
	}
	if cfg.Modules.Security == nil || !cfg.Modules.Security.Enabled {
		t.Errorf("expected security module to be enabled")
	}
	if cfg.Modules.Kubernetes != nil {
		t.Errorf("expected kubernetes module not enabled")
	}
}
