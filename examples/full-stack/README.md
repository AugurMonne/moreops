# Enterprise Full-Stack Example

This example demonstrates how MoreOps orchestrates an enterprise-ready microservice deployment with full observability, secret management, S3 storage, and multi-environment GitOps workflows.

## Features
- **Helm Deployment**: Scalable multi-replica deployment with ingress and resource limits.
- **Monitoring**: Prometheus Operator `ServiceMonitor` and Grafana deep-observability dashboard.
- **Secret Management**: External Secrets Operator connected to HashiCorp Vault.
- **Storage**: MinIO dev pod / Cloud S3 with automated daily backup schedule.
- **Multi-Environment**: Overrides for `dev`, `staging`, and `prod`.

## How to Run

1. **Inspect Plan**:
   ```bash
   moreops plan
   ```

2. **Test in Ephemeral Cluster**:
   ```bash
   moreops test --ephemeral
   ```

3. **Deploy to Target Environment**:
   ```bash
   moreops env switch prod
   moreops apply
   ```
