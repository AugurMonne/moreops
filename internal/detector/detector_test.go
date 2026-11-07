package detector

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectGoProject(t *testing.T) {
	tmpDir := t.TempDir()
	goMod := `module github.com/example/awesome-service

go 1.22

require github.com/gin-gonic/gin v1.9.1
`
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatal(err)
	}

	d := NewDetector()
	info, err := d.Detect(tmpDir)
	if err != nil {
		t.Fatalf("detect failed: %v", err)
	}

	if info.Language != "go" {
		t.Errorf("expected language go, got %s", info.Language)
	}
	if info.Framework != "gin" {
		t.Errorf("expected framework gin, got %s", info.Framework)
	}
	if info.Name != "awesome-service" {
		t.Errorf("expected name awesome-service, got %s", info.Name)
	}
}

func TestDetectNodeProject(t *testing.T) {
	tmpDir := t.TempDir()
	pkgJson := `{
  "name": "frontend-app",
  "dependencies": {
    "next": "^14.0.0"
  }
}`
	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(pkgJson), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "Dockerfile"), []byte("FROM node:alpine"), 0644); err != nil {
		t.Fatal(err)
	}

	d := NewDetector()
	info, err := d.Detect(tmpDir)
	if err != nil {
		t.Fatalf("detect failed: %v", err)
	}

	if info.Language != "node" {
		t.Errorf("expected language node, got %s", info.Language)
	}
	if info.Framework != "nextjs" {
		t.Errorf("expected framework nextjs, got %s", info.Framework)
	}
	if info.Name != "frontend-app" {
		t.Errorf("expected name frontend-app, got %s", info.Name)
	}
	if info.DefaultPort != 3000 {
		t.Errorf("expected default port 3000, got %d", info.DefaultPort)
	}
	if !info.HasDockerfile {
		t.Errorf("expected hasDockerfile to be true")
	}
}
