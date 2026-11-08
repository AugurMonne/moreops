# MoreOps Configuration Reference

The declarative configuration for MoreOps is stored in `moreops.yaml`.

## Core Sections

- `version`: Configuration schema version (e.g. `"1"`).
- `project`: Application metadata (name, version, description, programming language, framework).
- `environments`: Target environment mappings (`dev`, `staging`, `prod`) to Kubernetes clusters, namespaces, and clouds.
- `modules`: Configuration for selectable modules:
  - `containers`: Docker, containerd, docker-compose.
  - `kubernetes`: Workload strategy (Helm, Kustomize, Raw manifests), port, replicas, HPA, and Ingress controller (NGINX, Traefik, Caddy, HAProxy).
  - `database`: Database engines (PostgreSQL, MySQL, Redis), topologies (`shared` vs `dedicated`), and pre-upgrade migration Job hooks.
  - `messaging`: Message broker integration (RabbitMQ AMQP).
  - `monitoring`: Prometheus `ServiceMonitor`, Grafana dashboards, alert rules.
  - `security`: HashiCorp Vault (in-cluster dev or external), External Secrets Operator, RBAC, Trivy.
  - `storage`: S3/MinIO providers, Velero backup schedules, automated database backups.
  - `cicd`: GitLab CI, GitHub Actions, Jenkins (`Jenkinsfile`), Argo CD, OpenTofu, Ansible.

---

## Complete `moreops.yaml` Example

```yaml
version: "1"
project:
  name: my-service
  version: 0.1.0
  language: go
  framework: standard

environments:
  dev:
    cluster: k8s-dev
    namespace: dev
    cloud: local
  prod:
    cluster: k8s-prod
    namespace: prod
    cloud: yandex

modules:
  containers:
    enabled: true
    engine: docker
    compose: true

  kubernetes:
    enabled: true
    tool: helm # helm, kustomize, manifests
    port: 8080
    replicas: 2
    ingress: true
    ingress_class: nginx # nginx, traefik, caddy, haproxy, custom
    autoscaling:
      enabled: true
      min_replicas: 2
      max_replicas: 10
      cpu_percentage: 80

  database:
    enabled: true
    engine: postgres # postgres, mysql, redis
    topology: shared # "shared" (external/cloud) or "dedicated" (in-cluster)
    migration: true  # run pre-upgrade Helm migration Job
    host: "postgres.infra.svc.cluster.local"
    port: 5432
    name: "my_service_db"
    user: "my_service_user"

  messaging:
    enabled: true
    broker: rabbitmq
    host: "rabbitmq.infra.svc.cluster.local"
    port: 5672

  monitoring:
    enabled: true
    scrape_interval: 15s
    dashboards: true
    alert_rules: true

  security:
    enabled: true
    vault:
      enabled: true
      mode: in-cluster-dev # in-cluster-dev or external
    external_secrets:
      enabled: true

  storage:
    enabled: true
    provider: minio-dev
    bucket: app-data
    database_backup:
      enabled: true
      schedule: "0 2 * * *"

  cicd:
    enabled: true
    provider: gitlab # gitlab, github, jenkins
    gitops:
      enabled: true
      tool: argocd   # argocd, flux
      auto_sync: true
    iac:
      enabled: true
      tool: tofu
```

---

## Interactive Wizard Options Matrix (`moreops init`)

| Step / Option | Available Choices | Generated Output & Capabilities |
| :--- | :--- | :--- |
| **1. Metadata** | Project name and version | Base labels and metadata across `Chart.yaml` and manifests |
| **2. Target Environment** | Environment name (`dev`), cluster context, namespace | Sets up `environments` configuration and Helm values overlays |
| **3. Architecture Modules** | Components 1 to 6 (Containers, K8s, Monitoring, Security, Storage, CI/CD) | Enables corresponding generator modules |
| **4. Workload Strategy** | `[1]` Raw Manifests, `[2]` Helm (Recommended), `[3]` Kustomize | Production Helm chart in `charts/`, Kustomize overlays in `k8s/`, or raw manifests |
| **5. Replicas & HPA** | Initial replicas (default `2`) and HPA maximum (default `10`) | `HorizontalPodAutoscaler` manifest without GitOps replica count drift |
| **6. Ingress Controller** | `[1]` NGINX, `[2]` Traefik, `[3]` Caddy, `[4]` HAProxy, `[5]` None/Custom | Automated Ingress manifest with controller-specific annotations (`className`, TLS, entrypoints) |
| **7. Database Dependency** | `[1]` None, `[2]` PostgreSQL, `[3]` MySQL, `[4]` Redis | Docker Compose service with healthcheck, K8s SecretRef, and `DB_*` env variables |
| **8. Database Topology** | `[1]` Shared / Cloud DB (External), `[2]` Dedicated (In-Cluster) | Connects to existing cloud DB (Yandex, RDS) without overhead, or generates a local `StatefulSet` with 10Gi PVC |
| **9. Database Migrations** | Include migration Job (yes/no) | Manifest `migration-job.yaml` with Helm pre-upgrade hook preventing broken deployments |
| **10. RabbitMQ Broker** | Include message broker (yes/no) | `rabbitmq:3-management-alpine` in Compose with web UI (15672), K8s NetworkPolicy egress, and `RABBITMQ_*` envs |
