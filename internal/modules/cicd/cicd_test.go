package cicd

import (
	"context"
	"strings"
	"testing"

	"moreops/internal/core/config"
)

func TestCICDModuleGitHubAndArgoCD(t *testing.T) {
	cfg := config.NewDefaultConfig("delivery-srv")
	cfg.Modules.Kubernetes = &config.KubernetesModuleConfig{
		Enabled: true,
		Tool:    "helm",
	}
	cfg.Modules.CICD = &config.CICDModuleConfig{
		Enabled:    true,
		Provider:   "github",
		RunnerType: "cloud",
		GitOps: &config.GitOpsConfig{
			Enabled:  true,
			Tool:     "argocd",
			RepoURL:  "https://github.com/org/delivery-srv.git",
			Path:     "charts/delivery-srv",
			AutoSync: true,
		},
		IaC: &config.IaCConfig{
			Enabled: true,
			Tool:    "tofu",
		},
		Ansible: true,
	}

	mod := &Module{}
	plan, err := mod.Plan(context.Background(), cfg, "dev")
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}

	foundGitHubWorkflow := false
	foundArgoCD := false
	foundTerraform := false
	foundAnsible := false

	for _, fc := range plan.FileChanges {
		if strings.Contains(fc.Path, "ci-cd.yml") {
			foundGitHubWorkflow = true
			if !strings.Contains(fc.NewData, "runs-on: ubuntu-latest") || !strings.Contains(fc.NewData, "moreops validate") {
				t.Errorf("expected ubuntu-latest and moreops validate in github workflow")
			}
		}
		if strings.Contains(fc.Path, "argocd-application.yaml") {
			foundArgoCD = true
			if !strings.Contains(fc.NewData, "kind: Application") || !strings.Contains(fc.NewData, "namespace: argocd") {
				t.Errorf("expected ArgoCD Application manifest")
			}
		}
		if strings.Contains(fc.Path, "main.tf") {
			foundTerraform = true
			if !strings.Contains(fc.NewData, "app_name") {
				t.Errorf("expected app_name in terraform/main.tf")
			}
		}
		if strings.Contains(fc.Path, "playbook.yml") {
			foundAnsible = true
			if !strings.Contains(fc.NewData, "Bootstrap DevOps Host") {
				t.Errorf("expected Ansible playbook tasks")
			}
		}
	}

	if !foundGitHubWorkflow {
		t.Errorf("missing .github/workflows/ci-cd.yml")
	}
	if !foundArgoCD {
		t.Errorf("missing gitops/argocd-application.yaml")
	}
	if !foundTerraform {
		t.Errorf("missing terraform/main.tf")
	}
	if !foundAnsible {
		t.Errorf("missing ansible/playbook.yml")
	}
}

func TestCICDModuleGitLabAndSelfHosted(t *testing.T) {
	cfg := config.NewDefaultConfig("custom-ci-app")
	cfg.Modules.CICD = &config.CICDModuleConfig{
		Enabled:       true,
		Provider:      "gitlab",
		RunnerType:    "self-hosted",
		RunnerTag:     "k8s-runner",
		InstallRunner: true,
	}

	mod := &Module{}
	plan, err := mod.Plan(context.Background(), cfg, "dev")
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}

	foundGitLabCI := false
	foundRunnerDeploy := false

	for _, fc := range plan.FileChanges {
		if strings.Contains(fc.Path, ".gitlab-ci.yml") {
			foundGitLabCI = true
			if !strings.Contains(fc.NewData, "tags:\n    - k8s-runner") {
				t.Errorf("expected runner tag in .gitlab-ci.yml")
			}
		}
		if strings.Contains(fc.Path, "in-cluster-runner.yaml") {
			foundRunnerDeploy = true
			if !strings.Contains(fc.NewData, "k8s-runner") {
				t.Errorf("expected runner tag in in-cluster-runner.yaml")
			}
		}
	}

	if !foundGitLabCI {
		t.Errorf("missing .gitlab-ci.yml")
	}
	if !foundRunnerDeploy {
		t.Errorf("missing runners/in-cluster-runner.yaml")
	}
}

func TestCICDModuleJenkins(t *testing.T) {
	cfg := config.NewDefaultConfig("jenkins-app")
	cfg.Modules.CICD = &config.CICDModuleConfig{
		Enabled:   true,
		Provider:  "jenkins",
		RunnerTag: "k8s-nodes",
	}

	mod := &Module{}
	plan, err := mod.Plan(context.Background(), cfg, "prod")
	if err != nil {
		t.Fatalf("Plan failed: %v", err)
	}

	foundJenkinsfile := false
	for _, fc := range plan.FileChanges {
		if fc.Path == "Jenkinsfile" {
			foundJenkinsfile = true
			if !strings.Contains(fc.NewData, "pipeline {") || !strings.Contains(fc.NewData, "label 'k8s-nodes'") {
				t.Errorf("expected declarative pipeline with label k8s-nodes in Jenkinsfile")
			}
		}
	}

	if !foundJenkinsfile {
		t.Errorf("missing Jenkinsfile in plan")
	}
}

func TestCICDModuleDisabled(t *testing.T) {
	cfg := config.NewDefaultConfig("app")
	cfg.Modules.CICD = nil

	mod := &Module{}
	plan, err := mod.Plan(context.Background(), cfg, "dev")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan != nil {
		t.Errorf("expected nil plan when cicd is disabled")
	}
}
