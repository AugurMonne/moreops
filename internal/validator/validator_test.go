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

func TestValidatorHostAndPrivilegeRules(t *testing.T) {
	tmpDir := t.TempDir()

	hostPrivDep := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: privileged-app
spec:
  template:
    metadata:
      labels:
        app: privileged-app
    spec:
      hostNetwork: true
      hostPID: true
      volumes:
        - name: host-root
          hostPath:
            path: /
      containers:
        - name: app
          image: busybox
          securityContext:
            privileged: true
            runAsUser: 0
`
	if err := os.WriteFile(filepath.Join(tmpDir, "host-dep.yaml"), []byte(hostPrivDep), 0644); err != nil {
		t.Fatal(err)
	}

	val := NewValidator()
	issues, err := val.ValidateDirectory(tmpDir)
	if err != nil {
		t.Fatalf("validation failed: %v", err)
	}

	foundSec005 := false
	foundSec006 := false
	foundSec007 := false
	foundSec008 := false

	for _, issue := range issues {
		switch issue.RuleID {
		case "SEC005":
			foundSec005 = true
		case "SEC006":
			foundSec006 = true
		case "SEC007":
			foundSec007 = true
		case "SEC008":
			foundSec008 = true
		}
	}

	if !foundSec005 {
		t.Errorf("expected SEC005 (privileged: true) issue")
	}
	if !foundSec006 {
		t.Errorf("expected SEC006 (hostNetwork/hostPID) issue")
	}
	if !foundSec007 {
		t.Errorf("expected SEC007 (hostPath) issue")
	}
	if !foundSec008 {
		t.Errorf("expected SEC008 (runAsUser: 0) issue")
	}
}
