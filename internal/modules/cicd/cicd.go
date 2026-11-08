package cicd

import (
	"context"
	"fmt"
	"path/filepath"

	"moreops/internal/core/config"
	"moreops/internal/core/diff"
	"moreops/internal/core/plugin"
	"moreops/internal/integrations/jenkins"
)

// Module implements CI/CD (GitHub Actions, GitLab CI), GitOps (ArgoCD, Flux), OpenTofu/Terraform, and Ansible.
type Module struct{}

func init() {
	plugin.Register(&Module{})
}

func (m *Module) Name() string {
	return "cicd"
}

func (m *Module) Description() string {
	return "Manages CI/CD pipelines (GitHub Actions, GitLab CI), GitOps (Argo CD, Flux), OpenTofu, and Ansible"
}

func (m *Module) Plan(ctx context.Context, cfg *config.Config, envName string) (*plugin.PlanResult, error) {
	if cfg.Modules.CICD == nil || !cfg.Modules.CICD.Enabled {
		return nil, nil
	}

	env, _ := cfg.GetEnvironment(envName)
	namespace := env.Namespace
	if namespace == "" {
		namespace = "default"
	}

	appName := cfg.Project.Name
	ciCfg := cfg.Modules.CICD

	res := &plugin.PlanResult{
		ModuleName:  m.Name(),
		FileChanges: make([]diff.FileChange, 0),
		Summary:     fmt.Sprintf("CI/CD and automation configuration for %s", appName),
	}

	files := make(map[string]string)

	// Determine runner expression
	githubRunsOn := "ubuntu-latest"
	gitlabTags := ""
	if ciCfg.RunnerType == "self-hosted" {
		tag := ciCfg.RunnerTag
		if tag == "" {
			tag = "self-hosted"
		}
		githubRunsOn = fmt.Sprintf("[%s]", tag)
		gitlabTags = fmt.Sprintf("\n  tags:\n    - %s", tag)
	}

	// 1. GitHub Actions Pipeline
	if ciCfg.Provider == "github" {
		workflowFile := filepath.Join(".github", "workflows", "ci-cd.yml")
		files[workflowFile] = fmt.Sprintf(`name: CI/CD Pipeline

on:
  push:
    branches: [main, master]
  pull_request:
    branches: [main, master]

jobs:
  validate-and-test:
    name: Validate, Scan & Ephemeral Test
    runs-on: %s
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.22'

      - name: Build MoreOps CLI
        run: make build

      - name: Validate Manifests & Security
        run: ./bin/moreops validate

      - name: Security Vulnerability Scan (Trivy)
        run: ./bin/moreops scan fs

      - name: Ephemeral Cluster Verification
        run: ./bin/moreops test

  build-and-deploy:
    name: Build & Deploy
    needs: validate-and-test
    runs-on: %s
    if: github.ref == 'refs/heads/main' || github.ref == 'refs/heads/master'
    steps:
      - name: Checkout Code
        uses: actions/checkout@v4

      - name: Build Docker Image
        run: docker build -t %s:${{ github.sha }} .

      - name: MoreOps Dry-Run Plan
        run: ./bin/moreops plan --env %s

      - name: Apply Infrastructure Changes
        run: ./bin/moreops apply --yes --env %s
`, githubRunsOn, githubRunsOn, appName, envName, envName)
	}

	// 2. GitLab CI Pipeline
	if ciCfg.Provider == "gitlab" {
		gitlabFile := ".gitlab-ci.yml"
		files[gitlabFile] = fmt.Sprintf(`stages:
  - validate
  - security
  - test
  - build
  - deploy

default:
  image: golang:1.22-alpine%s

validate:
  stage: validate
  script:
    - apk add --no-cache make git
    - make build
    - ./bin/moreops validate

security-scan:
  stage: security
  image: aquasec/trivy:0.51.0
  script:
    - trivy fs --severity HIGH,CRITICAL .

build-image:
  stage: build
  image: docker:26.1
  services:
    - docker:26.1-dind
  script:
    - docker build -t %s:$CI_COMMIT_SHORT_SHA .
  only:
    - main
    - master

deploy:
  stage: deploy
  script:
    - make build
    - ./bin/moreops apply --yes --env %s
  only:
    - main
    - master
`, gitlabTags, appName, envName)
	}

	// 3. Jenkins Pipeline
	if ciCfg.Provider == "jenkins" {
		files["Jenkinsfile"] = jenkins.GenerateJenkinsfile(jenkins.PipelineOptions{
			AppName:     appName,
			Environment: envName,
			AgentLabel:  ciCfg.RunnerTag,
		})
	}

	// 3. In-Cluster Self-Hosted Runner Scaffold (if requested)
	if ciCfg.InstallRunner {
		runnerFile := filepath.Join("runners", "in-cluster-runner.yaml")
		files[runnerFile] = fmt.Sprintf(`apiVersion: apps/v1
kind: Deployment
metadata:
  name: %s-runner
  namespace: %s
  labels:
    app.kubernetes.io/name: ci-runner
spec:
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: ci-runner
  template:
    metadata:
      labels:
        app.kubernetes.io/name: ci-runner
    spec:
      containers:
        - name: runner
          image: my-docker-registry/ci-runner:latest
          env:
            - name: RUNNER_NAME
              value: "%s-k8s-runner"
            - name: RUNNER_TAGS
              value: "%s"
          resources:
            requests:
              cpu: 500m
              memory: 1Gi
            limits:
              cpu: 2000m
              memory: 4Gi
`, appName, namespace, appName, ciCfg.RunnerTag)
	}

	// 4. GitOps (Argo CD default or Flux CD)
	if ciCfg.GitOps != nil && ciCfg.GitOps.Enabled {
		if ciCfg.GitOps.Tool == "argocd" || ciCfg.GitOps.Tool == "" {
			argoFile := filepath.Join("gitops", "argocd-application.yaml")
			repoURL := ciCfg.GitOps.RepoURL
			if repoURL == "" {
				repoURL = "https://github.com/org/repo.git"
			}
			chartPath := ciCfg.GitOps.Path
			if chartPath == "" {
				chartPath = fmt.Sprintf("charts/%s", appName)
			}

			files[argoFile] = fmt.Sprintf(`apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: %s
  namespace: argocd
  labels:
    app.kubernetes.io/name: %s
  finalizers:
    - resources-finalizer.argocd.argoproj.io
spec:
  project: default
  source:
    repoURL: '%s'
    targetRevision: HEAD
    path: '%s'
    helm:
      valueFiles:
        - values.yaml
        - values-%s.yaml
  destination:
    server: 'https://kubernetes.default.svc'
    namespace: '%s'
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
      - ApplyOutOfSyncOnly=true
`, appName, appName, repoURL, chartPath, envName, namespace)
		} else if ciCfg.GitOps.Tool == "flux" {
			fluxFile := filepath.Join("gitops", "flux-kustomization.yaml")
			files[fluxFile] = fmt.Sprintf(`apiVersion: kustomize.toolkit.fluxcd.io/v1
kind: Kustomization
metadata:
  name: %s
  namespace: flux-system
spec:
  interval: 5m
  path: "./k8s"
  prune: true
  targetNamespace: %s
  sourceRef:
    kind: GitRepository
    name: %s-repo
`, appName, namespace, appName)
		}
	}

	// 5. OpenTofu / Terraform Baseline
	if ciCfg.IaC != nil && ciCfg.IaC.Enabled {
		tfMainFile := filepath.Join("terraform", "main.tf")
		files[tfMainFile] = fmt.Sprintf(`terraform {
  required_version = ">= 1.6.0"
  required_providers {
    local = {
      source  = "hashicorp/local"
      version = "~> 2.4"
    }
  }
}

# Baseline infrastructure parameters for %s
locals {
  app_name    = "%s"
  environment = "%s"
}

output "app_name" {
  value = local.app_name
}

output "environment" {
  value = local.environment
}
`, appName, appName, envName)

		tfVarsFile := filepath.Join("terraform", "variables.tf")
		files[tfVarsFile] = `variable "environment" {
  type        = string
  description = "Target deployment environment (dev, prod)"
  default     = "dev"
}
`
	}

	// 6. Ansible Host Bootstrap Playbook
	if ciCfg.Ansible {
		ansibleFile := filepath.Join("ansible", "playbook.yml")
		files[ansibleFile] = `---
- name: Bootstrap DevOps Host & Node Prerequisites
  hosts: all
  become: true
  tasks:
    - name: Update apt cache and install core utilities
      apt:
        update_cache: yes
        name:
          - curl
          - git
          - ufw
          - htop
          - ca-certificates
        state: present

    - name: Ensure sysctl params for Kubernetes networking
      sysctl:
        name: "{{ item.name }}"
        value: "{{ item.value }}"
        state: present
        reload: yes
      loop:
        - { name: 'net.bridge.bridge-nf-call-iptables', value: '1' }
        - { name: 'net.ipv4.ip_forward', value: '1' }

    - name: Disable swap for Kubernetes
      command: swapoff -a
      changed_when: false
`
		ansibleInv := filepath.Join("ansible", "inventory.ini")
		files[ansibleInv] = `[servers]
# 192.168.1.10 ansible_user=ubuntu
`
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
