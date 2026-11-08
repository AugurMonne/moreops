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

---

### Template Customization & Protection (`moreops template`)

- `moreops template list`: Catalog of customizable project templates.
- `moreops template eject [template-id | all] [--dir .moreops/templates]`: Eject built-in templates into `.moreops/templates/` for open-source project customization. MoreOps prioritizes local templates over built-in ones.
- **Custom Code Protection:** Adding `# moreops:ignore` or `# moreops:custom` in the first 10 lines of any generated file protects it from being overwritten by `moreops apply`.

---

### Secrets Management (`moreops secrets`)

- `moreops secrets init`: Setup local encryption key (`.moreops/secrets.key` with `0600` permissions) and `.sops.yaml`. Automatically adds key patterns to `.gitignore`.
- `moreops secrets template [name] [--backend vault|aws|gcp|azure]`: Generate ExternalSecret and SecretStore CRDs.
- `moreops secrets encrypt <file>`: Encrypt secrets file using SOPS or native AES-256-GCM.
- `moreops secrets decrypt <file>`: Decrypt secrets file safely.

---

### Diagnostics & Live Status (`moreops doctor`, `moreops status --live`)

- `moreops doctor`: Deep environment sanity, toolchain readiness, secret leak audit, and file permission checks.
- `moreops status [--live] [--env dev]`: Live cluster introspection (namespace, pods, restarts, and HPA).
