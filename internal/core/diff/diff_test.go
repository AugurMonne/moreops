package diff

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestComputeFileChangeCreate(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "new-file.txt")

	change, err := ComputeFileChange(targetPath, "hello world\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if change.Action != ActionCreate {
		t.Errorf("expected ActionCreate, got %s", change.Action)
	}

	diffText := UnifiedDiff(change, false)
	if !strings.Contains(diffText, "+++") || !strings.Contains(diffText, "+hello world") {
		t.Errorf("unexpected diff: %s", diffText)
	}
}

func TestComputeFileChangeUpdateAndNoop(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "existing.txt")

	if err := os.WriteFile(targetPath, []byte("line1\nline2\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Noop
	change, err := ComputeFileChange(targetPath, "line1\nline2\n")
	if err != nil {
		t.Fatal(err)
	}
	if change.Action != ActionNone {
		t.Errorf("expected ActionNone, got %s", change.Action)
	}

	// Update
	change, err = ComputeFileChange(targetPath, "line1\nline2 modified\n")
	if err != nil {
		t.Fatal(err)
	}
	if change.Action != ActionUpdate {
		t.Errorf("expected ActionUpdate, got %s", change.Action)
	}

	diffText := UnifiedDiff(change, false)
	if !strings.Contains(diffText, "-line2") || !strings.Contains(diffText, "+line2 modified") {
		t.Errorf("unexpected diff text: %s", diffText)
	}
}

func TestComputeFileChangeProtected(t *testing.T) {
	tmpDir := t.TempDir()
	targetPath := filepath.Join(tmpDir, "custom.yaml")

	content := "# moreops:ignore\napiVersion: apps/v1\nkind: Deployment\n"
	if err := os.WriteFile(targetPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	change, err := ComputeFileChange(targetPath, "apiVersion: apps/v1\nkind: Deployment\nreplicas: 5\n")
	if err != nil {
		t.Fatal(err)
	}

	if !change.Protected {
		t.Errorf("expected change to be protected by # moreops:ignore")
	}
	if change.Action != ActionNone {
		t.Errorf("expected ActionNone for protected file, got %s", change.Action)
	}

	diffText := UnifiedDiff(change, false)
	if !strings.Contains(diffText, "custom protected by # moreops:ignore") {
		t.Errorf("expected protected message in diff, got: %s", diffText)
	}

	// Test WriteFile does not overwrite protected file
	if err := WriteFile(targetPath, "overwritten content"); err != nil {
		t.Fatal(err)
	}
	readBack, _ := os.ReadFile(targetPath)
	if string(readBack) != content {
		t.Errorf("expected content to be preserved, got: %s", string(readBack))
	}
}
