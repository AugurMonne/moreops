# Configuration Reference

MoreOps reads its declarative configuration from `moreops.yaml`.

## Top-Level Fields

- `version`: Configuration schema version (e.g. `"1.0"`)
- `project`: Metadata including name, version, description, language, and framework
- `environments`: Target environment mappings (`dev`, `staging`, `prod`) with cluster, namespace, and cloud overrides
- `modules`: Selectable module settings:
  - `containers`: Docker, containerd, docker-compose
  - `kubernetes`: Helm chart, Kustomize, Raw manifests
  - `monitoring`: Prometheus, Grafana, ServiceMonitor, deep dashboards
  - `security`: Vault, External Secrets Operator, RBAC, Trivy
  - `storage`: S3/MinIO, Velero backups, Database CronJobs
  - `cicd`: Jenkins, GitHub Actions, GitLab CI, Argo CD, OpenTofu, Ansible
