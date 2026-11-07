package validator

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidatorSecurityAndConsistency(t *testing.T) {
	tmpDir := t.TempDir()

	// Insecure deployment missing probes, securityContext, and limits
	insecureDep := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: bad-app
spec:
  template:
    metadata:
      labels:
        app: bad-app
    spec:
      containers:
        - name: app
          image: nginx
`
	// Service that does not match bad-app labels
	mismatchedSvc := `apiVersion: v1
kind: Service
metadata:
  name: bad-service
spec:
  selector:
    app: wrong-label
  ports:
    - port: 80
`
	if err := os.WriteFile(filepath.Join(tmpDir, "dep.yaml"), []byte(insecureDep), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "svc.yaml"), []byte(mismatchedSvc), 0644); err != nil {
		t.Fatal(err)
	}

	val := NewValidator()
	issues, err := val.ValidateDirectory(tmpDir)
	if err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	foundSec001 := false
	foundSec002 := false
	foundCon001 := false

	for _, issue := range issues {
		if issue.RuleID == "SEC001" {
			foundSec001 = true
		}
		if issue.RuleID == "SEC002" {
			foundSec002 = true
		}
		if issue.RuleID == "CON001" {
			foundCon001 = true
		}
	}

	if !foundSec001 {
		t.Errorf("expected SEC001 (runAsNonRoot) issue")
	}
	if !foundSec002 {
		t.Errorf("expected SEC002 (allowPrivilegeEscalation) issue")
	}
	if !foundCon001 {
		t.Errorf("expected CON001 (Service selector mismatch) issue")
	}
}
