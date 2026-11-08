# MoreOps CLI Command Reference

Comprehensive reference for all MoreOps CLI commands.

---

### General Commands

- `moreops init`: Interactive onboarding wizard with language/framework detection.
- `moreops status`: Show project metadata, active environments, and manifests on disk.
- `moreops doctor`: Diagnostic report of installed CLI tools.
- `moreops version`: Display MoreOps version.
- `moreops completion [bash|zsh|fish]`: Generate shell autocompletion script.

---

### Manifests & Engine

- `moreops plan [--env <env>] [--config <path>]`: Generate diff without mutating files.
- `moreops apply [--env <env>] [--yes]`: Apply changes with unified diff preview and confirmation.
- `moreops generate [--env <env>] [--module <name>]`: Directly generate manifests for a specific module.
- `moreops validate [--config <path>] [--security] [--rbac]`: Static security linter and YAML/source code audit.

---

### Monitoring & Storage

- `moreops monitor status`: Check Prometheus ServiceMonitor and Grafana setup.
- `moreops storage check`: Inspect S3/MinIO provider and Velero backup configuration.

---

### Environments & Infrastructure

- `moreops env [list|switch|diff|add]`: Manage `dev`, `staging`, and `prod` targets.
- `moreops iac [plan|apply]`: OpenTofu / Terraform automation.
- `moreops chart [package|push]`: Package Helm charts and push to OCI registries.
- `moreops release [patch|minor|major]`: Automate SemVer bumping and changelog notes.
- `moreops test [--ephemeral]`: Run tests against ephemeral k3d clusters.
