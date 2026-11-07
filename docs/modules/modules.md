# MoreOps Modules Reference

Каждый модуль в MoreOps работает независимо и активируется в секции `modules` файла `moreops.yaml`.

---

## 1. Containers (`containers`)
- Генерация многоэтапного hardened `Dockerfile` под язык проекта (Go, Node.js, Python, Rust).
- Генерация `.dockerignore`.
- Генерация `docker-compose.yml` для локальной разработки со встроенным healthcheck.

## 2. Kubernetes (`kubernetes`)
- **Helm Chart (стандарт)**: создание структуры чарта `charts/<app>/` с шаблонами Deployment, Service, Ingress, HPA, NetworkPolicy и оверлеями values (`values-dev.yaml`, `values-prod.yaml`).
- **Kustomize**: структура `k8s/base/` и оверлеи окружений `k8s/overlays/dev/`, `k8s/overlays/prod/`.
- **Raw Manifests**: изолированные манифесты в `k8s/`.

## 3. Security & Secrets (`security`)
- **HashiCorp Vault**:
  - `in-cluster-dev`: автономный под Vault для локального тестирования секретов без внешних зависимостей.
  - `external`: подключение к внешнему Vault с Kubernetes аутентификацией.
- **External Secrets Operator (ESO)**: манифесты `SecretStore` и `ExternalSecret` для синхронизации секретов.
- **RBAC**: генерация строгого `ServiceAccount` с `automountServiceAccountToken: false`, `Role` и `RoleBinding`.

## 4. Monitoring & Observability (`monitoring`)
- **Prometheus Operator**: `ServiceMonitor` с 15-секундным сбором и пробросом метаданных пода и ноды.
- **PrometheusRule**: алерты на p99 latency (> 1s), ошибки 5xx (> 5%), перегрузку подов по CPU (> 85%), частые рестарты подов.
- **Grafana Dashboard**: `ConfigMap` с лейблом `grafana_dashboard: "1"` (Top подов по CPU/RAM, задержка запросов p50/p95/p99, сетевой трафик, RPS и ошибки).
- **Kube-Prometheus-Stack**: готовый `values.yaml` для официального Helm-чарта.

## 5. Storage & Backup (`storage`)
- **MinIO (Dev)**: автономный pod MinIO с S3 API (порт 9000) и веб-консолью.
- **S3 Провайдеры**: пресеты для Yandex Cloud Object Storage, VK Cloud, Selectel, AWS.
- **Velero**: манифесты `BackupStorageLocation` и `Schedule` с настраиваемым cron-расписанием.
- **Database Backup Workflows**: Kubernetes `CronJob` для безопасного дампирования PostgreSQL (`pg_dump`) или MySQL (`mysqldump`), компрессии и загрузки в S3.

## 6. CI/CD & Automation (`cicd`)
- **GitHub Actions**: `.github/workflows/ci-cd.yml` с поддержкой cloud и self-hosted раннеров.
- **GitLab CI**: `.gitlab-ci.yml`.
- **GitOps**: `Application` для Argo CD (по умолчанию) или Flux CD.
- **IaC**: базовые модули OpenTofu / Terraform.
- **Ansible**: плейбук начальной настройки хостов и нод.
