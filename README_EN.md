# MoreOps

[![Go Version](https://img.shields.io/github/go-mod/go-version/moreops/moreops)](https://golang.org)
[![License: Apache 2.0](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/tests-passing-brightgreen.svg)](tests/)

> **One CLI for your entire DevOps workflow.**  
> MoreOps does not reinvent tools — it orchestrates and manages proven standards: Docker, Kubernetes, Helm, Vault, S3, Prometheus, Jenkins, Argo CD, OpenTofu, and Ansible.

[Документация на русском](README.md) | **[English Documentation](README_EN.md)**

---

## Key Principles

1. **Selective & Independent Modularity**: Pick only the components you need (`1 2 4` or `14`).
2. **Mandatory Preview / Diff / Confirm**: Zero blind changes. Every command previews a unified diff and requires explicit confirmation.
3. **Multi-Cluster & Multi-Cloud**: First-class support for `dev`, `staging`, `prod` across multiple clusters and cloud providers.
4. **Integration Testing on Ephemeral Clusters**: Automatically spins up isolated `k3d` clusters for integration tests (`moreops test --ephemeral`).
5. **Secure by Default**: Non-root UID 10001, read-only root filesystems, drop capabilities, strict RBAC, automated security linter (`SEC001-SEC009`).
6. **Isolated Dev Mode**: In-cluster dev HashiCorp Vault and MinIO pods for zero-dependency local development.

---

## Quickstart

### 1. Build & Diagnostics

```bash
git clone https://github.com/moreops/moreops.git
cd moreops
make build

# Check available environment tools
./bin/moreops doctor
```

### 2. Interactive Initialization

```bash
./bin/moreops init
```
Automatically detects project language and framework, prompts for components, replicas, HPA, and generates declarative `moreops.yaml`.

### 3. Preview Changes (Dry-Run / Unified Diff)

```bash
./bin/moreops plan --env dev
```

### 4. Apply Changes (Confirm & Apply)

```bash
./bin/moreops apply --env dev
```

For CI/CD pipelines, skip interactive prompt:
```bash
./bin/moreops apply --yes --env prod
```

### 5. Validate & Security Linter

```bash
./bin/moreops validate
./bin/moreops validate --security
./bin/moreops validate --rbac
```

> 📖 **Step-by-step walkthrough**: see the [Manual Testing Guide](docs/en/manual-testing-guide.md).

---

## CLI Overview

- `moreops init`: Interactive project setup wizard with TUI arrow navigation (`↑`/`↓`), `Space` multi-select, and numeric jump keys.
- `moreops plan [--summary]`: Preview planned changes with unified diff or compact module summary table (`--summary`).
- `moreops apply [--yes]`: Apply changes with preview, interactive confirmation, and modular per-component synchronization status.
- `moreops diagnose [--namespace <ns>]`: Root-cause diagnostics for failing workloads, `CrashLoopBackOff`, `OOMKilled` exit 137, and warning events.
- `moreops validate`: Security, syntax, and RBAC policy linter.
- `moreops template [list|eject]`: Inspect and eject templates into `.moreops/templates/` for open-source customization.
- `moreops secrets [init|template|encrypt|decrypt]`: Manage secrets, generate ExternalSecret CRDs, and encrypt files.
- `moreops doctor [--fix]`: Comprehensive diagnostics for tools, runtimes, and secret leak detection with automated fix (`--fix`).
- `moreops status [--live]`: Show project status or query live Kubernetes cluster pods & HPA.
- `moreops chart [package|push]`: Package Helm charts (native pure Go fallback) and push to OCI registries.
- `moreops env [list|add|switch|diff]`: Multi-environment lifecycle management.
- `moreops iac [plan|apply]`: OpenTofu / Terraform workflow execution.
- `moreops cluster [list|create|delete]`: Ephemeral local k3d Kubernetes cluster manager.
- `moreops test [--ephemeral]`: Run validation tests against isolated ephemeral clusters.

---

## What You Can Choose & What You Get

| Wizard Option | Choices Available | Generated Output & Capabilities |
| :--- | :--- | :--- |
| **Workload Strategy** | `Helm` (Recommended), `Kustomize`, `Raw Manifests` | Full Helm chart (`charts/<app>/`) with values overlays, Kustomize overlays, or plain manifests |
| **Replicas & HPA** | Initial replica count + HPA max replicas | HorizontalPodAutoscaler manifest without GitOps flapping conflicts |
| **Ingress Controller** | `NGINX`, `Traefik`, `Caddy`, `HAProxy`, `Custom` | Ingress manifest with automated controller annotations (`className`, TLS, SSL redirect, router entrypoints) |
| **Database** | `None`, `PostgreSQL`, `MySQL`, `Redis` | Battle-tested stable images (`postgres:16-alpine`, `mysql:8.0`, `redis:7-alpine`), Compose service with healthchecks, K8s connection secrets and env (configurable via `.Values.database.image`) |
| **Database Topology** | `Shared / Cloud DB` vs `Dedicated In-Cluster` | Either connects to external cloud DB (RDS, Yandex) without cluster bloat, or deploys StatefulSet with 10Gi PVC, non-root security context, and engine-specific readiness checks (`pg_isready`, `mysqladmin ping`, `redis-cli ping`) |
| **DB Migrations** | Enable pre-upgrade migration Job (yes/no) | Kubernetes Job with Helm hook `pre-install,pre-upgrade` (`migrate/migrate:v4.17.0`) preventing broken releases |
| **RabbitMQ** | Enable AMQP message broker (yes/no) | `rabbitmq:3.13-management-alpine` in Compose with web UI (15672), K8s NetworkPolicy egress (port 5672), and env vars (configurable via `.Values.rabbitmq.image`) |
| **Hardened Runtimes** | Go, Node, Python, Rust, Shell | Modern vetted base images: Go on `distroless-debian12:nonroot`, Node 20 LTS, Python 3.11-slim, Alpine 3.20; CI/CD pinned to Docker 26.1 and Trivy 0.51.0 |
| **Custom Code Protection** | Add `# moreops:ignore` in any file | Prevents `moreops apply` from ever overwriting manual developer changes |
| **Plan Summary & Apply** | `moreops plan --summary`, `moreops apply` | High-level module changes table and modular per-component synchronization status |

- `moreops validate`: Validate YAML syntax, language code, and security rules (SEC001-SEC009).
- `moreops monitor status`: Check Prometheus Operator ServiceMonitor and alerts.
- `moreops storage check`: Check S3 / MinIO and Velero backup configuration.
- `moreops env [list|switch|diff|add]`: Multi-environment lifecycle management.
- `moreops chart [package|push]`: Package Helm charts (native pure Go fallback) and push to OCI registries.
- `moreops release [patch|minor|major]`: Automate SemVer bumping, Chart versioning, and CHANGELOG notes.
- `moreops test [--ephemeral]`: Run validation and verification on temporary k3d clusters.
