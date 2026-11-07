# Changelog

All notable changes to the MoreOps project are documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

---

## [0.1.0] - 2026-11-08

### Added
- **Core Engine (Phase 1)**: Interactive onboarding wizard (`moreops init`), project detection (Go, Node, Python, Rust), declarative configuration (`moreops.yaml`), unified diff preview (`moreops plan`), and safe apply (`moreops apply`).
- **Containers & Kubernetes (Phase 2)**: Multi-stage Dockerfile generation, Docker Compose, Helm chart scaffolding (`charts/<app>`), Kustomize overlays, Raw manifests, and ephemeral `k3d` testing clusters (`moreops cluster`, `moreops test --ephemeral`).
- **Security & Secrets (Phase 3)**: HashiCorp Vault (`in-cluster-dev` pod or external), External Secrets Operator (`SecretStore`, `ExternalSecret`), least-privilege RBAC generation, and vulnerability scanning via Trivy (`moreops scan fs`, `moreops scan image`).
- **Monitoring & Observability (Phase 4)**: Prometheus Operator `ServiceMonitor` (15s scrape interval, node & pod relabeling), `PrometheusRule` deep alerting (p99 latency, 5xx errors, pod load, restarts), and preconfigured Grafana Dashboard (`grafana_dashboard: "1"`).
- **Storage & Backup (Phase 5)**: In-cluster MinIO pod for dev, support for Russian and global cloud S3 providers (Yandex Cloud, VK Cloud, Selectel, AWS), Velero `BackupStorageLocation` & scheduled backups, and database backup CronJobs (PostgreSQL & MySQL).
- **Automation & CI/CD (Phase 6)**: GitHub Actions and GitLab CI pipelines with cloud and self-hosted runner support, Argo CD (default) and Flux CD GitOps sync, OpenTofu/Terraform baseline, and Ansible host bootstrap playbook.
- **DevOps Platform (Phase 7)**: Multi-environment management (`moreops env list`, `switch`, `diff`, `add`), automated release pipeline (`moreops release [patch|minor|major]`), and comprehensive documentation.
