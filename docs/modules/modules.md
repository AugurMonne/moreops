# MoreOps Modules Reference

Каждый модуль в MoreOps работает независимо и активируется в секции `modules` файла `moreops.yaml`. MoreOps не изобретает заново утилиты, а оркестрирует лучшие стандарты индустрии.

---

## 1. Containers (`containers`)
- **Hardened Multi-stage Dockerfile**: генерация минималистичного образа с distroless/alpine рантаймом под язык проекта (Go, Node.js, Python, Java, Rust, .NET).
- **UID/GID 10001 (Non-root)**: запуск контейнера от непривилегированного пользователя.
- **Docker Compose**: локальный стек (`docker-compose.yml`) со встроенными healthcheck-проверками.
- **`.dockerignore`**: исключение лишних файлов, артефактов компиляции и конфиденциальных `.env*`.

---

## 2. Kubernetes (`kubernetes`)
- **Helm Chart (дефолтный стандарт)**:
  - Структура чарта в `charts/<app>/` (Chart.yaml, values.yaml, templates/).
  - Шаблоны: Deployment, Service, Ingress, HPA, NetworkPolicy (default-deny egress/ingress).
  - Средовые оверлеи: `values-dev.yaml`, `values-prod.yaml`.
- **Kustomize**:
  - Базовые манифесты в `k8s/base/` с `kustomization.yaml`.
  - Оверлеи окружений в `k8s/overlays/dev/` и `k8s/overlays/prod/`.
- **Raw Manifests**:
  - Прямые манифесты для быстрого развертывания в `k8s/`.

---

## 3. Security & Secrets (`security`)
- **HashiCorp Vault**:
  - `in-cluster-dev`: автономный под Vault (режим dev) в кластере с преднастроенными health-пробами и hardened security context — идеален для локальной разработки без внешних зависимостей.
  - `external`: интеграция с корпоративным HashiCorp Vault через Kubernetes ServiceAccount auth.
- **External Secrets Operator (ESO)**:
  - Ресурс `SecretStore` для безопасной связи с Vault.
  - Ресурс `ExternalSecret` для декларативной материализации Kubernetes Secrets с интервалом обновления (например, `1h`).
- **RBAC**:
  - Создание изолированного `ServiceAccount` с `automountServiceAccountToken: false`.
  - Минимально необходимые `Role` и `RoleBinding`.
- **Trivy Scanner**:
  - Автоматическое сканирование уязвимостей контейнеров и зависимостей по шкале HIGH/CRITICAL.

---

## 4. Monitoring & Observability (`monitoring`)
- **Prometheus Operator**:
  - `ServiceMonitor` с 15-секундным сбором метрик (`/metrics`) и автоматическим пробросом меток пода (`pod`), ноды (`node`) и неймспейса (`namespace`).
- **Глубокие алерты (`PrometheusRule`)**:
  - Высокая задержка запросов (p99 latency > 1s).
  - Уровень ошибок HTTP 5xx (> 5%).
  - Перегрузка ноды/пода по CPU (> 85%) и памяти.
  - Частые перезапуски подов (CrashLoopBackOff).
- **Grafana Dashboard**:
  - Автоматический `ConfigMap` с лейблом `grafana_dashboard: "1"` для автоматического подхвата sidecar'ом Grafana.
  - Визуализация: распределение задержки p50/p95/p99, загрузка подов, сетевой трафик, RPS и HTTP-статусы.
- **Kube-Prometheus-Stack**:
  - Готовый файл `values.yaml` для официального чарта `prometheus-community/kube-prometheus-stack`.

---

## 5. Storage & Backup (`storage`)
- **In-cluster MinIO (Dev)**:
  - Автономный S3-совместимый под MinIO с веб-консолью для локального тестирования загрузки файлов и дампов.
- **S3 Провайдеры**:
  - **Yandex Cloud Object Storage** (`ru-central1`).
  - **VK Cloud Storage** (`ru-msk`).
  - **Selectel S3** (`ru-1`).
  - **AWS S3** (`us-east-1` и любые регионы).
- **Velero Backups**:
  - Ресурсы `BackupStorageLocation` и `Schedule` для создания снимков неймспейсов по расписанию (например, ежедневно в 2:00) с автоматической политикой хранения (TTL 30 дней).
- **Database Backup Workflows**:
  - Kubernetes `CronJob` для безопасного дампирования PostgreSQL (`pg_dump`) или MySQL (`mysqldump`), архивации (`gzip`) и потоковой выгрузки в S3-бакет.

---

## 6. CI/CD & Automation (`cicd`)
- **Jenkins**:
  - Декларативный `Jenkinsfile` со стадиями: Security Audit, Docker Build & Trivy Scan, Ephemeral Cluster Testing, GitOps Deploy.
- **GitHub Actions**:
  - Готовый пайплайн `.github/workflows/ci-cd.yml` с поддержкой cloud и self-hosted раннеров.
- **GitLab CI**:
  - `.gitlab-ci.yml` со стадиями validate, test, build, deploy.
- **In-Cluster Runner**:
  - Шаблон манифеста для развертывания CI-раннера внутри Kubernetes кластера.
- **GitOps**:
  - **Argo CD** (рекомендуемый дефолт): декларативный манифест `Application` с авто-синхронизацией (`automated.prune: true`, `selfHeal: true`).
  - **Flux CD**: манифесты `GitRepository` и `Kustomization` / `HelmRelease`.
- **Infrastructure as Code (IaC)**:
  - Базовые манифесты OpenTofu / Terraform (`main.tf`) для провижининга инфраструктуры.
- **Ansible**:
  - Готовый плейбук `playbook.yml` и инвентарь `inventory.ini` для базовой настройки хостов.
