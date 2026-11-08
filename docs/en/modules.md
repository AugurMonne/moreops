# MoreOps Modules Reference

Every module in MoreOps operates independently and can be toggled in the `modules` section of `moreops.yaml`. MoreOps does not reinvent tools — it orchestrates industry-proven standards with production-ready best practices.

---

## 1. Containers (`containers`)
- **Hardened Multi-stage Dockerfile**: Minimalist image generation using distroless or alpine runtimes tailored to your language (Go, Node.js, Python, Java, Rust, .NET).
- **UID/GID 65532 / 10001 (Non-root)**: Containers never run as root.
- **Custom Dockerfile Paths**: Full support for custom Dockerfile locations (`dockerfile: "services/api/Dockerfile"`, `context: "services/api"`), ideal for monorepos and multi-service repositories.
- **Preserve Existing Dockerfiles**: Option `keep_existing: true` keeps your current Dockerfile intact, generating only `.dockerignore` and configured `docker-compose.yml`.
- **Docker Compose**: Complete local development stack (`docker-compose.yml`) with automated health checks, database, and RabbitMQ.
- **`.dockerignore`**: Automatically prevents compilation artifacts, git history, and `.env*` files from entering builds.

---

## 2. Kubernetes (`kubernetes`)
- **Helm Chart (Default & Recommended)**:
  - Chart structure in `charts/<app>/` (Chart.yaml, values.yaml, values.schema.json, templates/).
  - Workload manifests: Deployment, Service, Ingress, HPA, PodDisruptionBudget, NetworkPolicy.
  - Multi-Environment Values Overlays:
    - **`dev`**: 1 replica, lightweight resources (50m/200m CPU, 64Mi/256Mi RAM), HPA disabled, Ingress `dev.<app>.local`, `LOG_LEVEL=debug`.
    - **`staging`**: 2 replicas, pre-production resources (100m/500m CPU, 128Mi/512Mi RAM), HPA 2-5 replicas, Ingress `staging.<app>.internal`, `LOG_LEVEL=info`.
    - **`prod`**: 3+ replicas, production limits (200m/1000m CPU, 256Mi/1Gi RAM), HPA 3-20 replicas, PodDisruptionBudget (`minAvailable: 1`), production TLS Ingress with cert-manager, `LOG_LEVEL=warn`.
  - JSON Schema validation (`values.schema.json`) for IDE auto-completion and `helm lint`.
- **Supported Ingress Controllers (What you choose & What you get)**:
  - `nginx` (**NGINX Ingress**, Default): `className: "nginx"`, auto-configured with SSL redirect annotations (`nginx.ingress.kubernetes.io/ssl-redirect: "true"`).
  - `traefik` (**Traefik Ingress**): `className: "traefik"`, auto-configured with Traefik router entrypoints (`traefik.ingress.kubernetes.io/router.entrypoints: web,websecure`, `traefik.ingress.kubernetes.io/router.tls: "true"`).
  - `caddy` (**Caddy Ingress**): `className: "caddy"`, auto-configured with automated TLS (`caddy.ingress.kubernetes.io/auto-https: "true"`).
  - `haproxy` (**HAProxy Ingress**): `className: "haproxy"`, auto-configured with HAProxy redirect annotations (`haproxy.org/ssl-redirect: "true"`).
  - `envoy` (**Envoy / Envoy Gateway**): `className: "envoy"`, auto-configured with Envoy annotations (`gateway.envoyproxy.io/secure: "true"`, `ingress.envoyproxy.io/ssl-redirect: "true"`), plus native Kubernetes Gateway API (`HTTPRoute`) template support.
  - `custom`: User-specified `ingressClassName` and custom annotations.
- **Kustomize**:
  - Base manifests in `k8s/base/` with `kustomization.yaml`.
  - Overlays in `k8s/overlays/dev/`, `k8s/overlays/staging/`, and `k8s/overlays/prod/`.
- **Raw Manifests**:
  - Direct Kubernetes manifests in `k8s/` configured with the selected Ingress controller.

---

## 3. Security & Secrets (`security`)
- **HashiCorp Vault**:
  - `in-cluster-dev`: Autonomous Vault dev pod inside the cluster for zero-dependency offline local development.
  - `external`: Production integration with enterprise Vault via Kubernetes ServiceAccount JWT tokens.
- **External Secrets Operator (ESO)**:
  - `SecretStore` resource configured for Vault, AWS Secrets Manager, GCP Secret Manager, or Azure Key Vault.
  - `ExternalSecret` resource for declarative synchronization into native Kubernetes Secrets.
- **Secrets Encryption (`moreops secrets`)**:
  - Local encryption key generation (`.moreops/secrets.key` with `0600` permissions) and `.sops.yaml`.
  - Encrypt/decrypt files using SOPS or native pure-Go AES-256-GCM fallback.
- **RBAC**:
  - Isolated `ServiceAccount` with `automountServiceAccountToken: false`.
  - Least-privilege `Role` and `RoleBinding`.

---

## 4. Monitoring & Observability (`monitoring`)
- **Prometheus Operator**:
  - `ServiceMonitor` scraping `/metrics` every 15 seconds.
  - `PrometheusRule` alerting on HighErrorRate, PodCrashLooping, HighMemoryUsage, HighCPUThrottling.
- **Grafana**:
  - Production dashboard JSON (`deep-observability.json`) providing instant graphs for CPU, RAM, HTTP rate, P99 latency, and pod restarts.
- **Loki & OpenTelemetry**:
  - Full-stack tracing and log aggregation support.

---

## 5. Storage & Backup (`storage`)
- **S3 & MinIO**:
  - Multi-cloud S3 baseline (Yandex Cloud Object Storage, VK Cloud, Selectel, AWS S3).
  - `minio-dev`: Lightweight in-cluster MinIO pod for offline local dev and S3 mocking.
- **Velero Cluster Backups**:
  - Automated Velero `Schedule` manifests for persistent volumes and cluster state.
- **Database Backup CronJob**:
  - Automated daily `pg_dump` or `mysqldump` CronJob uploading encrypted archives directly to S3.

---

## 6. CI/CD & Automation (`cicd`)
- **CI Pipelines**:
  - **GitLab CI**: Multi-stage `.gitlab-ci.yml` (test, validate, build, deploy).
  - **GitHub Actions**: Declarative workflow in `.github/workflows/ci-cd.yml`.
  - **Jenkins**: Declarative `Jenkinsfile`.
- **GitOps Reconciliation**:
  - **Argo CD** (Recommended default): Declarative `Application` manifest with automated prune and self-heal.
  - **Flux CD**: `GitRepository` and `HelmRelease` manifests.
- **Infrastructure as Code (IaC)**:
  - OpenTofu / Terraform root module (`main.tf`).
- **Ansible**:
  - Server bootstrap playbook (`playbook.yml`) and inventory (`inventory.ini`).

---

## 7. Databases (`database`)

Connect microservices to databases easily with support for local development, staging, and production topologies.

### Supported Engines:
- **`postgres` (PostgreSQL, Recommended)**: reliable stable `postgres:16-alpine`, port 5432.
- **`mysql` (MySQL)**: LTS image `mysql:8.0`, port 3306.
- **`redis` (Redis)**: lightweight `redis:7-alpine`, port 6379.

> All default images are defined in `.Values.database.image` and can be easily overridden.

### Cluster Topologies (What you choose & What you get):
1. **`shared` — Shared / Managed Cloud DB (Recommended for Production)**:
   * **Use Case:** Connects to an existing cloud database (Yandex Managed PostgreSQL, AWS RDS, GCP Cloud SQL) or centralized cluster DB.
   * **What You Get:**
     - Eliminates heavyweight in-cluster StatefulSets, saving cluster CPU/memory.
     - Connection parameters (`DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`) and SecretRef injected into Deployment.
     - Strict `NetworkPolicy` egress rules permitting TCP traffic to port 5432/3306 and HTTPS 443.
2. **`dedicated` — Dedicated In-Cluster Database**:
   * **Use Case:** Isolated QA environments, review apps, and local testing.
   * **What You Get:**
     - StatefulSet manifest (`charts/<app>/templates/database-statefulset.yaml`) with 10Gi PersistentVolumeClaim (dynamically configured for `postgres:16-alpine`, `mysql:8.0`, or `redis:7-alpine`) and Service.
     - Hardened Security Context: `runAsNonRoot: true`, `runAsUser: 999`, `allowPrivilegeEscalation: false`, `capabilities: drop: ["ALL"]`.
     - Automated readiness and liveness health checks (`pg_isready`, `mysqladmin ping`, or `redis-cli ping`).

### Pre-upgrade Database Migration Job (Helm Hook):
- **`migration: true`**:
  * **What You Get:** Manifest `charts/<app>/templates/migration-job.yaml` with annotations:
    ```yaml
    "helm.sh/hook": pre-install,pre-upgrade
    "helm.sh/hook-delete-policy": before-hook-creation,hook-succeeded
    ```
  * Before replacing application pods during an upgrade, Kubernetes executes this Job (`migrate/migrate`). If migrations fail, Helm aborts the upgrade, protecting production from database drift.

### Local Development (`docker-compose.yml`):
- Adds database container with persistent named volumes (`pgdata`, `mysqldata`, `redisdata`).
- Includes healthcheck and strict dependency `depends_on: { <db>: { condition: service_healthy } }` so the app only starts when the database is ready.

---

## 8. Message Broker (`messaging` / RabbitMQ)

Optional AMQP event bus for asynchronous microservices.

### What You Get:
1. **For Kubernetes:**
   - Values block in `values.yaml` linked to SecretRef.
   - Connection variables `RABBITMQ_HOST`, `RABBITMQ_PORT`, `RABBITMQ_USER`, `RABBITMQ_PASSWORD`, and AMQP URL injected into Deployment.
   - Egress rule in `NetworkPolicy` for TCP port `5672`.
2. **For Local Development (`docker-compose.yml`):**
   - Service `rabbitmq:3-management-alpine` exposing AMQP port `5672` and Web Management UI `15672`.
   - Healthcheck using `rabbitmq-diagnostics -q ping`.
   - Strict `depends_on: rabbitmq: { condition: service_healthy }`.
