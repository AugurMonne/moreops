# Workflows

MoreOps comes with preconfigured operational workflows in the `workflows/` directory:

- `workflows/application/deploy.yaml`: Complete build, validate, ephemeral test, and GitOps deploy flow.
- `workflows/monitoring/setup-observability.yaml`: Deploy Prometheus ServiceMonitors, alerts, and Grafana dashboards.
- `workflows/backup/disaster-recovery.yaml`: Backup scheduling and automated Velero snapshots to S3.
- `workflows/security/audit.yaml`: Trivy vulnerability scans and Kubernetes RBAC audits.
- `workflows/infrastructure/cluster-provision.yaml`: Spin up and tear down ephemeral `k3d` test clusters.
