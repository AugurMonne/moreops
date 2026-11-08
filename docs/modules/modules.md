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

---

## 7. Базы данных (`database`)

Модуль баз данных позволяет легко связать приложение с реляционными СУБД или кэшем, поддерживая как локальную разработку, так и облачные продакшен-топологии.

### Доступные движки:
- **`postgres` (PostgreSQL, Рекомендовано)**: образ `postgres:16-alpine`, порт `5432`.
- **`mysql` (MySQL)**: образ `mysql:8.0`, порт `3306`.
- **`redis` (Redis)**: образ `redis:7-alpine`, порт `6379`.

### Топологии развертывания (Что выбрать и что вы получите):
1. **`shared` — Общая / Managed Cloud DB (Рекомендовано для Production):**
   * **Когда выбирать:** Если у вас уже есть СУБД в облаке (Yandex Managed PostgreSQL, AWS RDS, VK Cloud DBaaS) или централизованный кластер баз данных в служебном неймспейсе (`infra`).
   * **Что вы получаете:**
     - В неймспейс приложения **не** развертывается тяжелый StatefulSet, экономя ресурсы кластера.
     - В `values.yaml` и `deployment.yaml` пробрасываются параметры подключения (`DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`), а пароль извлекается из `Secret` / `ExternalSecret`.
     - В `NetworkPolicy` автоматически добавляется разрешающее egress-правило на порт базы данных (5432/3306) и HTTPS (443).
2. **`dedicated` — Выделенная база в кластере (StatefulSet):**
   * **Когда выбирать:** Для локальной разработки, QA, демо-стендов или динамических Review-окружений, где микросервису требуется полностью изолированная БД.
   * **Что вы получаете:**
     - Манифест `charts/<app>/templates/database-statefulset.yaml` со `StatefulSet`, PVC объемом `10Gi` на базе `postgres:16-alpine` и `Service` ClusterIP.
     - Hardened Security Context: запуск от `runAsNonRoot: true`, `runAsUser: 999`, запрет эскалации привилегий (`allowPrivilegeEscalation: false`), сброс всех capabilities (`drop: ["ALL"]`).
     - Встроенные пробы готовности и жизни через утилиту `pg_isready`.

### Автоматические миграции схемы (Database Migration Job):
- **`migration: true`**:
  * **Что вы получаете:** Создание манифеста `charts/<app>/templates/migration-job.yaml` с аннотацией хука Helm:
    ```yaml
    "helm.sh/hook": pre-install,pre-upgrade
    "helm.sh/hook-delete-policy": before-hook-creation,hook-succeeded
    ```
  * Перед выкаткой новой версии сервиса Kubernetes запускает Job миграций (`migrate/migrate`). Если миграция упала, Helm прерывает деплой, защищая работающий прод от падений.

### Локальная разработка (`docker-compose.yml`):
- Автоматическое добавление контейнера базы данных с персистентным томом (`pgdata` / `mysqldata` / `redisdata`).
- Проверка работоспособности (`healthcheck`) и строгая зависимость `depends_on: { <db>: { condition: service_healthy } }` — сервис приложения гарантированно ждет готовности базы перед стартом.

---

## 8. Очереди сообщений (`messaging` / RabbitMQ)

Модуль очередей сообщений опционален и предназначен для микросервисов с асинхронным взаимодействием.

### Что вы получаете при включении RabbitMQ:
1. **Для Kubernetes:**
   - В `values.yaml` прописывается блок `rabbitmq` с хостом, портом и ссылкой на Secret.
   - В `deployment.yaml` инжектируются переменные `RABBITMQ_HOST`, `RABBITMQ_PORT`, `RABBITMQ_USER`, `RABBITMQ_PASSWORD` и URL вида `amqp://...`.
   - В `NetworkPolicy` чарта добавляется egress-правило для порта `5672` (AMQP).
2. **Для локальной разработки (`docker-compose.yml`):**
   - Контейнер `rabbitmq:3-management-alpine` с открытыми портами:
     - `5672` — протокол AMQP для приложения.
     - `15672` — Web UI панели управления (Management Dashboard).
   - Healthcheck через `rabbitmq-diagnostics -q ping`.
   - Блок `depends_on: rabbitmq: condition: service_healthy` в контейнере приложения.
   - Персистентный том `rabbitmqdata` для сохранения очередей при перезапусках.
