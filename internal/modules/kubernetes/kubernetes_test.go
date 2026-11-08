package kubernetes

import (
	"context"
	"strings"
	"testing"

	"moreops/internal/core/config"
)

func TestKubernetesPlanWithDatabaseAndRabbitMQ(t *testing.T) {
	m := &Module{}
	cfg := &config.Config{
		Project: config.ProjectMeta{
			Name: "my-web-api",
		},
		Environments: map[string]config.EnvironmentConfig{
			"dev": {
				Cluster:   "k8s-dev",
				Namespace: "staging",
			},
		},
		Modules: config.ModulesConfig{
			Kubernetes: &config.KubernetesModuleConfig{
				Enabled:  true,
				Tool:     "helm",
				Port:     8080,
				Replicas: 2,
				Autoscaling: &config.AutoscalingConfig{
					Enabled:     true,
					MinReplicas: 2,
					MaxReplicas: 8,
				},
			},
			Database: &config.DatabaseModuleConfig{
				Enabled:   true,
				Engine:    "postgres",
				Topology:  "dedicated",
				Migration: true,
			},
			Messaging: &config.MessagingModuleConfig{
				Enabled: true,
				Broker:  "rabbitmq",
				Host:    "rabbitmq.infra.svc.cluster.local",
				Port:    5672,
			},
		},
	}

	plan, err := m.Plan(context.Background(), cfg, "dev")
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}

	var foundMigrationJob, foundStatefulSet, foundDeployment, foundValues bool
	for _, fc := range plan.FileChanges {
		if strings.HasSuffix(fc.Path, "migration-job.yaml") {
			foundMigrationJob = true
			if !strings.Contains(fc.NewData, "helm.sh/hook") || !strings.Contains(fc.NewData, "pre-upgrade") {
				t.Error("expected pre-upgrade Helm hook in migration-job.yaml")
			}
		}
		if strings.HasSuffix(fc.Path, "database-statefulset.yaml") {
			foundStatefulSet = true
			if !strings.Contains(fc.NewData, "kind: StatefulSet") {
				t.Error("expected StatefulSet in database-statefulset.yaml")
			}
		}
		if strings.HasSuffix(fc.Path, "deployment.yaml") {
			foundDeployment = true
			if !strings.Contains(fc.NewData, "DB_HOST") || !strings.Contains(fc.NewData, "RABBITMQ_HOST") {
				t.Error("expected DB and RabbitMQ env vars in deployment.yaml")
			}
		}
		if strings.HasSuffix(fc.Path, "values.yaml") {
			foundValues = true
			if !strings.Contains(fc.NewData, "database:") || !strings.Contains(fc.NewData, "rabbitmq:") {
				t.Error("expected database and rabbitmq blocks in values.yaml")
			}
		}
	}

	if !foundMigrationJob {
		t.Error("migration-job.yaml not found in plan")
	}
	if !foundStatefulSet {
		t.Error("database-statefulset.yaml not found in plan")
	}
	if !foundDeployment {
		t.Error("deployment.yaml not found in plan")
	}
	if !foundValues {
		t.Error("values.yaml not found in plan")
	}
}

func TestKubernetesPlanWithTraefikAndCaddyIngress(t *testing.T) {
	m := &Module{}
	cfg := &config.Config{
		Project: config.ProjectMeta{
			Name: "traefik-api",
		},
		Environments: map[string]config.EnvironmentConfig{
			"dev": {
				Cluster:   "k8s-dev",
				Namespace: "default",
			},
		},
		Modules: config.ModulesConfig{
			Kubernetes: &config.KubernetesModuleConfig{
				Enabled:      true,
				Tool:         "helm",
				Port:         8080,
				Replicas:     2,
				Ingress:      true,
				IngressClass: "traefik",
			},
		},
	}

	plan, err := m.Plan(context.Background(), cfg, "dev")
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}

	var valuesContent string
	for _, fc := range plan.FileChanges {
		if strings.HasSuffix(fc.Path, "values.yaml") {
			valuesContent = fc.NewData
		}
	}

	if !strings.Contains(valuesContent, "className: \"traefik\"") {
		t.Errorf("expected className traefik in values.yaml, got:\n%s", valuesContent)
	}
	if !strings.Contains(valuesContent, "traefik.ingress.kubernetes.io/router.entrypoints") {
		t.Errorf("expected traefik annotations in values.yaml, got:\n%s", valuesContent)
	}
}

func TestKubernetesPlanWithEnvoyIngress(t *testing.T) {
	m := &Module{}
	cfg := &config.Config{
		Project: config.ProjectMeta{
			Name: "envoy-api",
		},
		Environments: map[string]config.EnvironmentConfig{
			"dev": {
				Cluster:   "k8s-dev",
				Namespace: "default",
			},
		},
		Modules: config.ModulesConfig{
			Kubernetes: &config.KubernetesModuleConfig{
				Enabled:      true,
				Tool:         "helm",
				Port:         8080,
				Replicas:     2,
				Ingress:      true,
				IngressClass: "envoy",
			},
		},
	}

	plan, err := m.Plan(context.Background(), cfg, "dev")
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}

	var valuesContent, foundHttpRoute bool
	for _, fc := range plan.FileChanges {
		if strings.HasSuffix(fc.Path, "values.yaml") {
			if strings.Contains(fc.NewData, "className: \"envoy\"") && strings.Contains(fc.NewData, "gateway.envoyproxy.io/secure") {
				valuesContent = true
			}
		}
		if strings.HasSuffix(fc.Path, "httproute.yaml") {
			foundHttpRoute = true
		}
	}

	if !valuesContent {
		t.Error("expected envoy className and annotations in values.yaml")
	}
	if !foundHttpRoute {
		t.Error("expected httproute.yaml in templates")
	}
}

func TestKubernetesPlanMultiEnvironments(t *testing.T) {
	m := &Module{}
	cfg := &config.Config{
		Project: config.ProjectMeta{
			Name: "multi-env-app",
		},
		Environments: map[string]config.EnvironmentConfig{
			"dev": {
				Cluster:   "local-ctx",
				Namespace: "default",
			},
			"staging": {
				Cluster:   "staging-ctx",
				Namespace: "staging",
			},
			"prod": {
				Cluster:   "prod-ctx",
				Namespace: "production",
			},
		},
		Modules: config.ModulesConfig{
			Kubernetes: &config.KubernetesModuleConfig{
				Enabled:  true,
				Tool:     "helm",
				Port:     8080,
				Replicas: 2,
			},
		},
	}

	plan, err := m.Plan(context.Background(), cfg, "prod")
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}

	var foundDev, foundStaging, foundProd, foundPDB bool
	for _, fc := range plan.FileChanges {
		if strings.HasSuffix(fc.Path, "values-dev.yaml") {
			foundDev = true
			if !strings.Contains(fc.NewData, "replicaCount: 1") || !strings.Contains(fc.NewData, "ENVIRONMENT") {
				t.Error("expected dev overrides in values-dev.yaml")
			}
		}
		if strings.HasSuffix(fc.Path, "values-staging.yaml") {
			foundStaging = true
			if !strings.Contains(fc.NewData, "replicaCount: 2") || !strings.Contains(fc.NewData, "staging.multi-env-app.internal") {
				t.Error("expected staging overrides in values-staging.yaml")
			}
		}
		if strings.HasSuffix(fc.Path, "values-prod.yaml") {
			foundProd = true
			if !strings.Contains(fc.NewData, "replicaCount: 3") || !strings.Contains(fc.NewData, "podDisruptionBudget:") {
				t.Error("expected prod overrides with PDB in values-prod.yaml")
			}
		}
		if strings.HasSuffix(fc.Path, "poddisruptionbudget.yaml") {
			foundPDB = true
		}
	}

	if !foundDev {
		t.Error("values-dev.yaml not found in plan")
	}
	if !foundStaging {
		t.Error("values-staging.yaml not found in plan")
	}
	if !foundProd {
		t.Error("values-prod.yaml not found in plan")
	}
	if !foundPDB {
		t.Error("poddisruptionbudget.yaml not found in plan")
	}
}
