package security

import (
	"context"
	"strings"
	"testing"

	"moreops/internal/core/config"
)

func TestSecurityModulePlan(t *testing.T) {
	cfg := config.NewDefaultConfig("my-secure-app")
	cfg.Modules.Kubernetes = &config.KubernetesModuleConfig{
		Enabled: true,
		Tool:    "helm",
	}
	cfg.Modules.Security = &config.SecurityModuleConfig{
		Enabled: true,
		Vault: &config.VaultConfig{
			Enabled: true,
			Mode:    "in-cluster-dev",
		},
		ExternalSecrets: &config.ExternalSecretsConfig{
			Enabled:         true,
			RefreshInterval: "1h",
			SecretStoreName: "vault-store",
		},
		RBAC: &config.RBACConfig{
			Mode: "strict",
		},
	}

	mod := &Module{}
	plan, err := mod.Plan(context.Background(), cfg, "dev")
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}

	if plan == nil {
		t.Fatalf("expected non-nil plan")
	}

	foundRBAC := false
	foundVaultDev := false
	foundSecretStore := false
	foundExternalSecret := false

	for _, fc := range plan.FileChanges {
		if strings.Contains(fc.Path, "rbac.yaml") {
			foundRBAC = true
			if !strings.Contains(fc.NewData, "automountServiceAccountToken: false") {
				t.Errorf("expected automountServiceAccountToken: false in ServiceAccount")
			}
		}
		if strings.Contains(fc.Path, "vault-dev.yaml") {
			foundVaultDev = true
		}
		if strings.Contains(fc.Path, "secretstore.yaml") {
			foundSecretStore = true
			if !strings.Contains(fc.NewData, "kind: SecretStore") {
				t.Errorf("expected kind: SecretStore")
			}
		}
		if strings.Contains(fc.Path, "externalsecret.yaml") {
			foundExternalSecret = true
			if !strings.Contains(fc.NewData, "kind: ExternalSecret") {
				t.Errorf("expected kind: ExternalSecret")
			}
		}
	}

	if !foundRBAC {
		t.Errorf("missing rbac.yaml in plan")
	}
	if !foundVaultDev {
		t.Errorf("missing vault-dev.yaml in plan")
	}
	if !foundSecretStore {
		t.Errorf("missing secretstore.yaml in plan")
	}
	if !foundExternalSecret {
		t.Errorf("missing externalsecret.yaml in plan")
	}
}

func TestSecurityModuleDisabled(t *testing.T) {
	cfg := config.NewDefaultConfig("plain-app")
	cfg.Modules.Security = nil

	mod := &Module{}
	plan, err := mod.Plan(context.Background(), cfg, "dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan != nil {
		t.Errorf("expected nil plan when security module is nil/disabled")
	}
}
