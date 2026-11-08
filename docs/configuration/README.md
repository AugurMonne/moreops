# Справочник конфигурации MoreOps

Декларативное описание проекта хранится в файле `moreops.yaml`.

## Основные секции

- `version`: Версия схемы конфигурации (например, `"1"`).
- `project`: Метаданные сервиса (имя, версия, описание, язык программирования, фреймворк).
- `environments`: Карта целевых окружений (`dev`, `staging`, `prod`) с привязкой к Kubernetes-контекстам/кластерам, неймспейсам и облакам.
- `modules`: Настройки независимых модулей:
  - `containers`: Docker, containerd, docker-compose.
  - `kubernetes`: Стратегия деплоя (Helm-чарт, Kustomize, Raw манифесты), порт, реплики, HPA автоскейлинг.
  - `monitoring`: Prometheus Operator `ServiceMonitor`, пороги алертов, Grafana дашборды.
  - `security`: HashiCorp Vault (in-cluster dev или external), External Secrets Operator, RBAC, Trivy.
  - `storage`: S3/MinIO, расписание снимков Velero, бэкапы баз данных.
  - `cicd`: Jenkins (`Jenkinsfile`), GitHub Actions, GitLab CI, Argo CD, OpenTofu, Ansible.
  - `database`: Интеграция с СУБД (PostgreSQL, MySQL, Redis), выбор топологии (`shared` или `dedicated`) и хуки миграций.
  - `messaging`: Очереди сообщений на базе RabbitMQ (AMQP 5672 + Web UI 15672).

---

## Пример полного файла `moreops.yaml`

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
    autoscaling:
      enabled: true
      min_replicas: 2
      max_replicas: 10
      cpu_percentage: 80

  database:
    enabled: true
    engine: postgres # postgres, mysql, redis
    topology: shared # "shared" (внешняя/cloud DB) или "dedicated" (в кластере)
    migration: true  # запуск Helm pre-upgrade Job с миграциями
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
      mode: in-cluster-dev # in-cluster-dev или external
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

## Опции выбора в интерактивном мастере (`moreops init`)

При запуске мастера `moreops init` вам последовательно предлагаются следующие шаги:

| Шаг | Что можно выбрать | Что вы получаете в результате |
| :--- | :--- | :--- |
| **1. Метаданные** | Имя и версия проекта | Базовые теги в `Chart.yaml`, метки `app.kubernetes.io/name` |
| **2. Окружение** | Имя среды (`dev`), контекст K8s и namespace | Настройка секции `environments` и параметров Helm values |
| **3. Модули** | Компоненты от 1 до 6 (Containers, K8s, Monitoring, Security, Storage, CI/CD) | Включение соответствующих генераторов манифестов |
| **4. Стратегия K8s** | `[1]` Raw Manifests, `[2]` Helm (Recommended), `[3]` Kustomize | Полноценный Helm-чарт в `charts/`, Kustomize-оверлеи или чистые манифесты |
| **5. Реплики и HPA** | Число реплик (дефолт `2`) и макс. порог HPA (дефолт `10`) | Настройка горизонтального автоскейлинга без конфликтов с GitOps |
| **6. База данных** | `[1]` None, `[2]` PostgreSQL, `[3]` MySQL, `[4]` Redis | Docker Compose сервис с healthcheck, переменные подключения в K8s |
| **7. Топология БД** | `[1]` Shared / Cloud DB (Внешняя), `[2]` Dedicated (В кластере) | Либо экономия ресурсов и подключение к Cloud DB, либо готовый StatefulSet с PVC 10Gi |
| **8. Миграции БД** | Включать ли Job миграций (yes/no) | Helm-манифест `migration-job.yaml` с pre-upgrade хуком |
| **9. RabbitMQ** | Включать ли брокер сообщений (yes/no) | Сервис `rabbitmq:3-management-alpine` в Compose, сетевые политики и env в K8s |

