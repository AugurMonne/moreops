package security

import (
	"context"
	"fmt"
	"path/filepath"

	"moreops/internal/core/config"
	"moreops/internal/core/diff"
	"moreops/internal/core/plugin"
)

// Module implements Vault, External Secrets Operator, and RBAC management.
type Module struct{}

func init() {
	plugin.Register(&Module{})
}

func (m *Module) Name() string {
	return "security"
}

func (m *Module) Description() string {
	return "Manages HashiCorp Vault, External Secrets Operator (ESO), and RBAC"
}

func (m *Module) Plan(ctx context.Context, cfg *config.Config, envName string) (*plugin.PlanResult, error) {
	if cfg.Modules.Security == nil || !cfg.Modules.Security.Enabled {
		return nil, nil
	}

	env, _ := cfg.GetEnvironment(envName)
	namespace := env.Namespace
	if namespace == "" {
		namespace = "default"
	}

	appName := cfg.Project.Name
	secCfg := cfg.Modules.Security

	res := &plugin.PlanResult{
		ModuleName:  m.Name(),
		FileChanges: make([]diff.FileChange, 0),
		Summary:     fmt.Sprintf("Security configuration (Vault, ESO, RBAC) for %s", appName),
	}

	// Determine output base directory based on k8s strategy
	isHelm := cfg.Modules.Kubernetes != nil && cfg.Modules.Kubernetes.Tool == "helm"
	var outDir string
	if isHelm {
		outDir = filepath.Join("charts", appName, "templates")
	} else {
		outDir = filepath.Join("k8s", "security")
	}

	files := make(map[string]string)

	// 1. RBAC (Strict Least-Privilege)
	if secCfg.RBAC != nil && secCfg.RBAC.Mode == "strict" {
		rbacFile := filepath.Join(outDir, "rbac.yaml")
		files[rbacFile] = fmt.Sprintf(`apiVersion: v1
kind: ServiceAccount
metadata:
  name: %s
  namespace: %s
  labels:
    app.kubernetes.io/name: %s
automountServiceAccountToken: false
---
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: %s-role
  namespace: %s
  labels:
    app.kubernetes.io/name: %s
rules:
  - apiGroups: [""]
    resources: ["configmaps"]
    verbs: ["get", "list", "watch"]
---
apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  name: %s-rolebinding
  namespace: %s
  labels:
    app.kubernetes.io/name: %s
subjects:
  - kind: ServiceAccount
    name: %s
    namespace: %s
roleRef:
  kind: Role
  name: %s-role
  apiGroup: rbac.authorization.k8s.io
`, appName, namespace, appName, appName, namespace, appName, appName, namespace, appName, appName, namespace, appName)
	}

	// 2. In-cluster Dev Vault (for local testing without external dependencies)
	vaultAddress := "http://vault.vault.svc:8200"
	if secCfg.Vault != nil {
		if secCfg.Vault.Address != "" {
			vaultAddress = secCfg.Vault.Address
		}
		if secCfg.Vault.Mode == "in-cluster-dev" {
			vaultDevFile := filepath.Join(outDir, "vault-dev.yaml")
			files[vaultDevFile] = fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: vault-dev
  namespace: %s
  labels:
    app.kubernetes.io/name: vault-dev
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: vault-dev
  template:
    metadata:
      labels:
        app.kubernetes.io/name: vault-dev
    spec:
      containers:
        - name: vault
          image: hashicorp/vault:1.16
          ports:
            - containerPort: 8200
              name: http
          securityContext:
            allowPrivilegeEscalation: false
            capabilities:
              drop:
                - ALL
              add:
                - IPC_LOCK
          env:
            - name: VAULT_DEV_ROOT_TOKEN_ID
              value: "root-dev-token"
            - name: VAULT_DEV_LISTEN_ADDRESS
              value: "0.0.0.0:8200"
          resources:
            requests:
              cpu: 50m
              memory: 64Mi
            limits:
              cpu: 250m
              memory: 256Mi
          livenessProbe:
            httpGet:
              path: /v1/sys/health?standbyok=true
              port: 8200
            initialDelaySeconds: 5
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /v1/sys/health?standbyok=true
              port: 8200
            initialDelaySeconds: 5
            periodSeconds: 10
---
apiVersion: v1
kind: Service
metadata:
  name: vault-dev
  namespace: %s
spec:
  type: ClusterIP
  selector:
    app.kubernetes.io/name: vault-dev
  ports:
    - name: http
      port: 8200
      targetPort: 8200
`, namespace, namespace)
			vaultAddress = fmt.Sprintf("http://vault-dev.%s.svc:8200", namespace)
		}
	}

	// 3. External Secrets Operator (SecretStore + ExternalSecret)
	if secCfg.ExternalSecrets != nil && secCfg.ExternalSecrets.Enabled {
		storeName := secCfg.ExternalSecrets.SecretStoreName
		if storeName == "" {
			storeName = "vault-backend"
		}
		refreshInterval := secCfg.ExternalSecrets.RefreshInterval
		if refreshInterval == "" {
			refreshInterval = "1h"
		}

		secretStoreFile := filepath.Join(outDir, "secretstore.yaml")
		files[secretStoreFile] = fmt.Sprintf(`apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: %s
  namespace: %s
spec:
  provider:
    vault:
      server: "%s"
      path: "secret"
      version: "v2"
      auth:
        tokenSecretRef:
          name: vault-token
          key: token
`, storeName, namespace, vaultAddress)

		externalSecretFile := filepath.Join(outDir, "externalsecret.yaml")
		files[externalSecretFile] = fmt.Sprintf(`apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: %s-secrets
  namespace: %s
spec:
  refreshInterval: "%s"
  secretStoreRef:
    name: %s
    kind: SecretStore
  target:
    name: %s-env
    creationPolicy: Owner
  data:
    - secretKey: DB_PASSWORD
      remoteRef:
        key: %s/credentials
        property: password
    - secretKey: API_KEY
      remoteRef:
        key: %s/credentials
        property: api_key
`, appName, namespace, refreshInterval, storeName, appName, appName, appName)
	}

	for path, content := range files {
		fc, err := diff.ComputeFileChange(path, content)
		if err != nil {
			return nil, err
		}
		res.FileChanges = append(res.FileChanges, fc)
	}

	return res, nil
}

func (m *Module) Apply(ctx context.Context, plan *plugin.PlanResult) error {
	return plugin.ExecutePlan(plan)
}
