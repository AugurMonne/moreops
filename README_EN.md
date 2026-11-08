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

- `moreops init`: Interactive project setup wizard with auto-detection.
- `moreops plan`: Preview planned infrastructure changes with unified diff.
- `moreops apply`: Apply changes with interactive preview and confirmation (`--yes`).
- `moreops generate`: Generate manifests for a specific module (`--module monitoring`).
- `moreops status`: Show project status, environments, and on-disk manifests.
- `moreops validate`: Validate YAML syntax, language code, and security rules (SEC001-SEC009).
- `moreops monitor status`: Check Prometheus Operator ServiceMonitor and alerts.
- `moreops storage check`: Check S3 / MinIO and Velero backup configuration.
- `moreops env [list|switch|diff|add]`: Multi-environment lifecycle management.
- `moreops chart [package|push]`: Package Helm charts (native pure Go fallback) and push to OCI registries.
- `moreops release [patch|minor|major]`: Automate SemVer bumping, Chart versioning, and CHANGELOG notes.
- `moreops test [--ephemeral]`: Run validation and verification on temporary k3d clusters.
