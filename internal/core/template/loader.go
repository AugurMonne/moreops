package template

import (
	"fmt"
	"os"
	"path/filepath"
)

// TemplateInfo contains metadata for an ejectable template.
type TemplateInfo struct {
	ID          string
	Category    string
	Description string
	DefaultPath string
	Content     string
}

// BuiltinTemplates returns the catalog of customizable templates in MoreOps.
func BuiltinTemplates() map[string]TemplateInfo {
	return map[string]TemplateInfo{
		"docker/go": {
			ID:          "docker/go",
			Category:    "Containers",
			Description: "Multi-stage hardened Go Dockerfile with non-root distroless runtime",
			DefaultPath: "templates/docker/go/Dockerfile",
			Content: `# Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /src
RUN apk add --no-cache ca-certificates git
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-w -s" -o /bin/app ./cmd/...

# Final minimal secure runtime
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /bin/app /app/server
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/app/server"]
`,
		},
		"docker/compose": {
			ID:          "docker/compose",
			Category:    "Containers",
			Description: "Local development Docker Compose stack",
			DefaultPath: "templates/compose/docker-compose.yml",
			Content: `version: "3.8"
services:
  app:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      - ENV=development
      - PORT=8080
`,
		},
		"kubernetes/deployment": {
			ID:          "kubernetes/deployment",
			Category:    "Kubernetes",
			Description: "Kubernetes Deployment manifest template with security context & probes",
			DefaultPath: "templates/kubernetes/workloads/deployment.yaml",
			Content: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Name }}
  namespace: {{ .Namespace }}
  labels:
    app.kubernetes.io/name: {{ .Name }}
spec:
  {{- if not .AutoscalingEnabled }}
  replicas: {{ .Replicas }}
  {{- end }}
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ .Name }}
  template:
    metadata:
      labels:
        app.kubernetes.io/name: {{ .Name }}
    spec:
      securityContext:
        runAsNonRoot: true
        runAsUser: 65532
        runAsGroup: 65532
        fsGroup: 65532
      containers:
        - name: {{ .Name }}
          image: "{{ .Image }}"
          imagePullPolicy: IfNotPresent
          securityContext:
            allowPrivilegeEscalation: false
            readOnlyRootFilesystem: true
            capabilities:
              drop: ["ALL"]
          ports:
            - name: http
              containerPort: {{ .Port }}
              protocol: TCP
          livenessProbe:
            httpGet:
              path: /healthz
              port: http
            initialDelaySeconds: 5
            periodSeconds: 10
          readinessProbe:
            httpGet:
              path: /healthz
              port: http
            initialDelaySeconds: 5
            periodSeconds: 10
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 500m
              memory: 512Mi
`,
		},
		"cicd/gitlab": {
			ID:          "cicd/gitlab",
			Category:    "CI/CD",
			Description: "GitLab CI multi-stage pipeline (test, validate, build, deploy)",
			DefaultPath: "templates/cicd/gitlab-ci/.gitlab-ci.yml",
			Content: `stages:
  - test
  - validate
  - build
  - deploy

lint_and_test:
  stage: test
  image: golang:1.22-alpine
  script:
    - go test -v ./...

security_audit:
  stage: validate
  image: alpine:3.20
  script:
    - moreops validate --security

build_image:
  stage: build
  image: docker:26.1
  services:
    - docker:26.1-dind
  script:
    - docker build -t $CI_REGISTRY_IMAGE:$CI_COMMIT_SHA .
`,
		},
		"cicd/github": {
			ID:          "cicd/github",
			Category:    "CI/CD",
			Description: "GitHub Actions workflow for continuous integration and delivery",
			DefaultPath: "templates/cicd/github-actions/ci-cd.yml",
			Content: `name: CI/CD Pipeline

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Run Tests
        run: go test -v ./...
      - name: Security Audit
        run: moreops validate --security
`,
		},
		"cicd/jenkins": {
			ID:          "cicd/jenkins",
			Category:    "CI/CD",
			Description: "Jenkins declarative pipeline with stages for audit and container build",
			DefaultPath: "templates/cicd/jenkins/Jenkinsfile",
			Content: `pipeline {
    agent any
    stages {
        stage('Test & Audit') {
            steps {
                sh 'moreops validate'
            }
        }
        stage('Build') {
            steps {
                sh 'docker build -t app:${BUILD_NUMBER} .'
            }
        }
    }
}
`,
		},
	}
}

// Resolve looks for a user-customized template in .moreops/templates/<id>
// or project-relative templates/<id>, falling back to defaultContent.
func Resolve(templateID, defaultContent string) string {
	customPath := filepath.Join(".moreops", "templates", templateID)
	if data, err := os.ReadFile(customPath); err == nil && len(data) > 0 {
		return string(data)
	}

	repoPath := filepath.Join("templates", templateID)
	if data, err := os.ReadFile(repoPath); err == nil && len(data) > 0 {
		return string(data)
	}

	bt := BuiltinTemplates()
	if info, ok := bt[templateID]; ok {
		if data, err := os.ReadFile(info.DefaultPath); err == nil && len(data) > 0 {
			return string(data)
		}
		if defaultContent == "" {
			return info.Content
		}
	}

	return defaultContent
}

// Eject writes built-in templates into targetDir (default: .moreops/templates/)
// so the user can easily inspect and customize them.
func Eject(targetDir string, templateIDs ...string) ([]string, error) {
	if targetDir == "" {
		targetDir = filepath.Join(".moreops", "templates")
	}

	bt := BuiltinTemplates()
	selected := make(map[string]TemplateInfo)

	if len(templateIDs) == 0 || (len(templateIDs) == 1 && templateIDs[0] == "all") {
		selected = bt
	} else {
		for _, id := range templateIDs {
			info, ok := bt[id]
			if !ok {
				return nil, fmt.Errorf("unknown template '%s' (run 'moreops template list' to see available)", id)
			}
			selected[id] = info
		}
	}

	ejected := make([]string, 0, len(selected))
	for id, info := range selected {
		destPath := filepath.Join(targetDir, id)
		dir := filepath.Dir(destPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}

		content := info.Content
		if data, err := os.ReadFile(info.DefaultPath); err == nil && len(data) > 0 {
			content = string(data)
		}

		if err := os.WriteFile(destPath, []byte(content), 0644); err != nil {
			return nil, fmt.Errorf("failed writing template %s: %w", destPath, err)
		}
		ejected = append(ejected, destPath)
	}

	return ejected, nil
}
