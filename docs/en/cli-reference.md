# MoreOps CLI Command Reference

Comprehensive reference for all MoreOps CLI commands.

---

### General Commands

- `moreops init`: Interactive onboarding wizard with language/framework detection. Features arrow navigation (`↑`/`↓` or `j`/`k`), checkbox multi-select (`Space`), quick numeric jump keys (`1-N`), and fallback for non-TTY pipelines.
- `moreops status [--live]`: Show project metadata, active modules, and query live Kubernetes cluster pods & HPA.
- `moreops doctor [--fix]`: Comprehensive environment, toolchain, and secret hygiene check. `--fix` (`-f`) automatically fixes file permissions (0600) and `.gitignore`.
- `moreops diagnose [--env <env>] [--namespace <ns>]` (aliases: `troubleshoot`, `debug`): Root cause analysis for failing pods (`CrashLoopBackOff`, `OOMKilled` exit 137, `ImagePullBackOff`, restarts) with actionable fix suggestions.
- `moreops version`: Display MoreOps version.
- `moreops completion [bash|zsh|fish]`: Generate shell autocompletion script.

---

### Manifests & Engine

- `moreops plan [--env <env>] [--config <path>] [--summary]`: Generate diff without mutating files. Flag `--summary` (`-s`) outputs a high-level module status table instead of raw YAML diffs.
- `moreops apply [--env <env>] [--yes]`: Apply changes with preview, interactive confirmation, and modular per-component synchronization progress. Use `--yes` for CI/CD pipelines.
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
