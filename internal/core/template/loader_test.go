package template

import (
	"os"
	"testing"
)

func TestBuiltinTemplates(t *testing.T) {
	templates := BuiltinTemplates()
	if len(templates) == 0 {
		t.Fatal("expected builtin templates catalog to be non-empty")
	}

	if _, ok := templates["docker/go"]; !ok {
		t.Error("expected docker/go template to exist")
	}
}

func TestResolveAndEject(t *testing.T) {
	tempDir := t.TempDir()

	ejected, err := Eject(tempDir, "docker/go")
	if err != nil {
		t.Fatalf("Eject failed: %v", err)
	}
	if len(ejected) != 1 {
		t.Fatalf("expected 1 ejected file, got %d", len(ejected))
	}

	content, err := os.ReadFile(ejected[0])
	if err != nil {
		t.Fatalf("failed reading ejected file: %v", err)
	}

	if len(content) == 0 {
		t.Error("ejected template content is empty")
	}

	// Verify fallback when no custom file exists
	resolved := Resolve("nonexistent", "fallback")
	if resolved != "fallback" {
		t.Errorf("expected fallback, got %s", resolved)
	}
}

func TestEjectAll(t *testing.T) {
	tempDir := t.TempDir()
	ejected, err := Eject(tempDir, "all")
	if err != nil {
		t.Fatalf("Eject all failed: %v", err)
	}
	if len(ejected) != len(BuiltinTemplates()) {
		t.Errorf("expected %d ejected files, got %d", len(BuiltinTemplates()), len(ejected))
	}
}
