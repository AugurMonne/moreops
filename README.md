# MoreOps

<p align="center">
  <strong>Universal DevOps & Infrastructure Orchestrator written in Go.</strong><br>
  <em>One CLI to manage, secure, monitor, and automate your entire DevOps stack without reinventing the wheel.</em>
</p>

<p align="center">
  <a href="#core-principles">Principles</a> •
  <a href="#quickstart">Quickstart</a> •
  <a href="#architecture--modules">Modules</a> •
  <a href="#cli-reference">CLI Reference</a> •
  <a href="#documentation">Docs</a> •
  <a href="LICENSE">License</a>
</p>

---

## Overview

**MoreOps** — это модульный DevOps-оркестратор нового поколения. Вместо того чтобы создавать очередные кастомные аналоги Prometheus, HashiCorp Vault, Terraform или Argo CD, **MoreOps управляет ими**, связывая разрозненные индустриальные инструменты в единую декларативную платформу.

### Ключевые архитектурные принципы

- **Оркестрация, а не изобретение велосипедов**: Управление проверенными инструментами индустрии (Helm, Kustomize, Vault, ESO, Prometheus, Velero, OpenTofu, Ansible, Argo CD).
- **Селективность и модульность**: Вы выбираете только нужные компоненты. Никакого навязанного монолитного стека.
- **Гарантия Preview / Diff / Confirm**: Ни одно изменение не вносится вслепую. Каждая операция проходит стадии генерации плана, расчета unified diff и явного подтверждения.
- **Декларативность и GitOps**: Единый файл `moreops.yaml` описывает окружения и модули, готовый для версионирования в Git.
- **Multi-Cluster & Multi-Cloud**: Поддержка изолированных сред (`dev`, `staging`, `prod`) с привязкой к разным Kubernetes контекстам/кластерам, неймспейсам и облакам (Yandex Cloud, VK Cloud, Selectel, AWS).
- **Тестирование на временных кластерах**: Быстрый запуск эфемерных `k3d` кластеров для E2E верификации перед деплоем в прод (`moreops test --ephemeral`).
- **Безопасность по умолчанию (Secure by default)**: Pod Security Standards (non-root UID 10001, read-only rootfs, drop capabilities), шифрование секретов в Vault, least-privilege RBAC и встроенный сканер Trivy.

---

## Архитектура и модули (Phases 1–7)

MoreOps полностью покрывает все 7 этапов DevOps-пайплайна:

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                 MoreOps CLI                                 │
│         (Init Wizard, Multi-Environment Engine, Diff & Apply Engine)        │
└──────┬────────────┬─────────────┬─────────────┬─────────────┬─────────────┬─┘
       │            │             │             │             │             │
┌──────v──────┐┌────v───────┐┌────v───────┐┌────v───────┐┌────v───────┐┌────v───────┐
│ Containers  ││ Kubernetes ││  Security  ││ Monitoring ││  Storage   ││ Automation  │
│  & Docker   ││Helm/Kustom.││ Vault/ESO  ││Prometheus  ││ MinIO/S3   ││CI/CD/GitOps │
│ Compose     ││ k3d Ephem. ││ RBAC/Trivy ││Grafana/Loki││Velero/DB   ││Tofu/Ansible │
└─────────────┘└────────────┘└────────────┘└────────────┘└────────────┘└────────────┘
```

1. **Phase 1 — Core Engine**: Интерактивный Wizard (`moreops init`), автодетект проекта (Go, Node, Python, Rust), парсинг мульти-выбора (`14` / `1 4`), движок Unified Diff.
2. **Phase 2 — Containers & Kubernetes**: Multi-stage Dockerfile, Docker Compose, Helm-чарты (`charts/<app>`), Kustomize, Raw манифесты, управление локальными `k3d` кластерами (`moreops cluster`).
3. **Phase 3 — Security & Secrets**: Автономный in-cluster dev Vault pod или подключение к External Vault, External Secrets Operator, строгий RBAC, сканирование уязвимостей через Trivy (`moreops scan`).
4. **Phase 4 — Monitoring & Observability**: Prometheus Operator `ServiceMonitor` (15s скрейп с relabeling подов и нод), `PrometheusRule` (p99 latency, 5xx ошибки, перегрузка CPU, рестарты), преднастроенный Grafana Dashboard (`grafana_dashboard: "1"`).
5. **Phase 5 — Storage & Backup**: In-cluster MinIO pod, российские и глобальные S3 провайдеры (Yandex Cloud, VK Cloud, Selectel, AWS), Velero BSL и расписания бэкапов, CronJob для дампов баз данных (PostgreSQL / MySQL).
6. **Phase 6 — Automation & CI/CD**: Jenkins (`Jenkinsfile`), GitHub Actions и GitLab CI пайплайны (cloud и self-hosted раннеры), GitOps синхронизация через **Argo CD** (дефолт) и Flux CD, OpenTofu/Terraform (`moreops iac plan/apply`), Ansible плейбук настройки хостов.
7. **Phase 7 — DevOps Platform**: Управление мульти-окружениями (`moreops env list/switch/diff/add`), автоматизация версий и релизов (`moreops release [patch|minor|major]`), полная документация.

---

## Quickstart

### 1. Установка и сборка

```bash
git clone https://github.com/moreops/moreops.git
cd moreops
make build
```

Проверьте доступность инструментов в вашей системе:

```bash
./bin/moreops doctor
```

### 2. Инициализация проекта

Запустите интерактивный мастер со стрелочками (`↑`/`↓`, `Space`, `Enter`):

```bash
./bin/moreops init
```

Мастер автоматически:
- Определит язык и стек проекта (Go, Node.js, Python, Rust).
- Предложит выбрать целевые окружения (`dev`, `staging`, `prod`).
- Проверит наличие существующего Dockerfile: предложит сохранить ваш файл (`keep`) или сгенерировать новый, а также настроить путь для монорепозиториев (`services/api/Dockerfile`).
- Позволит выбрать нужные модули (Containers, Kubernetes, Database, RabbitMQ, CI/CD, Security, Monitoring).
- Создаст декларативный `moreops.yaml`.

### 3. Предпросмотр изменений (Preview / Diff)

Сгенерируйте и посмотрите цветной unified diff без изменения файлов:

```bash
# Для окружения разработки
./bin/moreops plan --env dev

# Компактная сводка изменений
./bin/moreops plan --summary

# Для продакшена
./bin/moreops plan --env prod
```

### 4. Применение конфигурации (Confirm & Apply)

```bash
./bin/moreops apply --env dev
```

MoreOps выведет diff и запросит подтверждение:
`Do you want to perform these actions? (yes/no): `

Для CI/CD автоматизации передайте флаг `--yes`:

```bash
./bin/moreops apply --yes --env prod
```

### 5. Валидация и проверка безопасности

```bash
# Проверка синтаксиса и Security Linter (SEC001-SEC009, REL001, CON001)
./bin/moreops validate

# Сканирование файловой системы и манифестов через Trivy
./bin/moreops scan fs

# Запуск тестов на временном k3d кластере
./bin/moreops test --ephemeral
```

### 6. Самообновление (moreops upgrade)

Чтобы обновить MoreOps до самого свежего релиза с GitHub в один клик:

```bash
moreops upgrade

# Или проверить наличие обновлений без скачивания
moreops upgrade --check
```

> 📖 **Пошаговое руководство по проверке вручную**: см. [Manual Testing Guide](docs/getting-started/manual-testing-guide.md).

---

## Что можно выбрать и что вы получаете при выборе

| Пункт мастера / Опция | Доступные варианты | Что вы получаете в результате |
| :--- | :--- | :--- |
| **Целевые окружения** | `dev`, `staging`, `prod` | Индивидуальные values-файлы (`values-dev.yaml`, `values-staging.yaml`, `values-prod.yaml`) и Kustomize-оверлеи с тонкой настройкой реплик, ресурсов, HPA, PDB и TLS |
| **Расположение Dockerfile** | `Keep existing`, `Overwrite`, `Custom path` | Бережная защита существующего Dockerfile (`keep`), поддержка путей для монорепозиториев (`services/api/Dockerfile`) |
| **Стратегия Kubernetes** | `Helm` (Рекомендовано), `Kustomize`, `Raw Manifests` | Полноценный Helm-чарт в `charts/<app>/` со схемой валидации `values.schema.json`, Kustomize-структура в `k8s/` или чистые манифесты |
| **Реплики и HPA** | Начальное число подов + HPA максимум | `HorizontalPodAutoscaler` без конфликтов с GitOps (поле `replicas` в Deployment опускается при включенном HPA) |
| **Ingress-контроллер** | `NGINX`, `Traefik`, `Caddy`, `HAProxy`, `Envoy`, `Custom` | Автоматическая генерация Ingress манифеста с аннотациями выбранного контроллера (`className`, TLS, SSL redirect, router entrypoints, Gateway API `HTTPRoute` для Envoy) |
| **База данных** | `None`, `PostgreSQL`, `MySQL`, `Redis` | Проверенные стабильные образы (`postgres:16-alpine`, `mysql:8.0`, `redis:7-alpine`), готовый сервис в `docker-compose.yml` с healthcheck, проброс переменных `DB_*` и SecretRef в Kubernetes (версия переопределяется в `.Values.database.image`) |
| **Топология БД** | `Shared / Cloud DB` vs `Dedicated In-Cluster` | Подключение к существующей Cloud DB (Yandex/AWS) без оверхеда, либо локальный `StatefulSet` с PVC 10Gi в неймспейсе с проверками готовности (`pg_isready`, `mysqladmin ping`, `redis-cli ping`) |
| **Миграции БД** | Включение Job миграций (yes/no) | Helm pre-upgrade Job (`migrate/migrate:v4.17.0`) — блокирует деплой при ошибках миграций, защищая прод |
| **RabbitMQ** | Включение брокера (yes/no) | `rabbitmq:3.13-management-alpine` в Compose с Web UI (15672), сетевые политики NetworkPolicy (порт 5672) и переменные `RABBITMQ_*` |
| **Базовые рантаймы** | Go, Node, Python, Rust, Shell | Проверенные современные образы: Go на `distroless-debian12:nonroot`, Node 20 LTS, Python 3.11-slim, Alpine 3.20; CI/CD на Docker 26.1 и Trivy 0.51.0 |
| **Управление секретами** | `moreops secrets [init\|template\|encrypt]` | Локальный ключ AES-256-GCM, генерация `ExternalSecret`/`SecretStore` CRD, безопасное шифрование файлов |
| **Кастомизация шаблонов** | `moreops template [list\|eject]` | Извлечение шаблонов в `.moreops/templates/` для open-source доработок; защита ручного кода тегом `# moreops:ignore` |
| **Диагностика окружения** | `moreops doctor` | Проверка Docker daemon, kubectl, helm, git, прав доступа к файлам (0600) и аудит утечек токенов в конфигах |
| **Сводка плана и Apply** | `moreops plan --summary`, `moreops apply` | Высокоуровневая сводка изменений по модулям и покомпонентная индикация статуса применения |
| **Статус живого кластера** | `moreops status --live` | Опрос активного контекста K8s: реальное состояние подов, перезапуски, утилизация HPA |


---

## Пример конфигурации `moreops.yaml`

```yaml
version: "1"
project:
  name: payment-service
  version: 1.0.0
  language: go
  framework: gin

environments:
  dev:
    cluster: k3d-dev-cluster
    namespace: payments-dev
    cloud: local
  prod:
    cluster: prod-k8s-cluster
    namespace: payments-prod
    cloud: yandex

modules:
  containers:
    enabled: true
    engine: docker
    image: payment-service:latest
    compose: true

  kubernetes:
    enabled: true
    tool: helm
    port: 8080
    replicas: 3
    ingress: true

  security:
    enabled: true
    vault:
      enabled: true
      mode: in-cluster-dev # или "external"
    external_secrets:
      enabled: true
      refresh_interval: 1h
    rbac:
      mode: strict

  monitoring:
    enabled: true
    prometheus: true
    grafana: true
    alertmanager: true
    scrape_interval: 15s
    dashboards: true
    alert_rules: true

  storage:
    enabled: true
    provider: yandex # yandex, vkcloud, selectel, minio-dev, aws
    bucket: payment-backups
    velero:
      enabled: true
      schedule: "0 2 * * *"
    database_backup:
      enabled: true
      engine: postgres
      schedule: "0 3 * * *"

  cicd:
    enabled: true
    provider: github
    runner_type: cloud
    gitops:
      enabled: true
      tool: argocd
      auto_sync: true
    iac:
      enabled: true
      tool: tofu
    ansible: true
```

---

## Справочник команд CLI

| Команда | Описание |
| :--- | :--- |
| `moreops init` | Интерактивный мастер создания проекта с TUI-навигацией стрелками (`↑`/`↓`), `Space`, `Enter` |
| `moreops doctor [--fix]` | Диагностика установленных DevOps инструментов и авто-исправление прав/gitignore (`--fix`) |
| `moreops diagnose` | Интеллектуальный траблшутинг упавших подов, `CrashLoopBackOff`, `OOMKilled` и Warning-событий |
| `moreops plan [--summary]` | Предпросмотр изменений с unified diff или компактной сводной таблицей (`--summary`) |
| `moreops apply [--yes]` | Безопасное применение изменений с покомпонентным прогрессом синхронизации |
| `moreops status [--live]` | Статус модулей или опрос живого Kubernetes кластера (состояние подов, HPA) |
| `moreops validate` | Аудит конфигурации и встроенный Security Linter (CIS Benchmark) |
| `moreops scan [fs\|image]` | Сканирование уязвимостей через Trivy (с Docker fallback) |
| `moreops template [list\|eject]` | Извлечение встроенных шаблонов в `.moreops/templates/` для кастомизации |
| `moreops secrets [init\|encrypt\|decrypt]` | Управление ключами AES-256-GCM и шифрование секретов |
| `moreops cluster [list\|create\|delete]` | Управление локальными кластерами `k3d` |
| `moreops test [--ephemeral]` | Тестирование манифестов на временных k8s-кластерах |
| `moreops env [list\|switch\|diff\|add]` | Управление несколькими окружениями (dev, staging, prod) |
| `moreops iac [plan\|apply]` | Управление IaC через OpenTofu / Terraform |
| `moreops release [patch\|minor\|major]` | Автоматический релиз, бамп SemVer и генерация CHANGELOG |
| `moreops version` | Показать версию MoreOps CLI |

---

## Разработка

Для сборки и тестирования MoreOps можно использовать как `make`, так и `task`:

```bash
# Использование Makefile
make test      # Запуск всех модульных тестов
make lint      # Запуск golangci-lint
make build     # Сборка bin/moreops

# Использование Taskfile (кроссплатформенно)
task test
task lint
task build
task release-dry-run
```

Сборка и запуск через Docker:

```bash
docker build -t moreops:latest .
docker run --rm -v $(pwd):/workspace moreops:latest plan
```

---

## Документация

Полная документация находится в каталоге [`docs/`](docs/):

- [Быстрый старт (Quickstart Guide)](docs/getting-started/quickstart.md)
- [Архитектура и ключевые принципы](docs/architecture/overview.md)
- [Полный справочник команд CLI](docs/commands/cli-reference.md)
- [Руководство по модулям](docs/modules/modules.md)
- [Стандарты безопасности (Security Best Practices)](docs/security/best-practices.md)

---

## Лицензия

Проект распространяется под лицензией [MIT](LICENSE).  
Политика безопасности описана в [SECURITY.md](SECURITY.md).  
Правила участия в разработке — в [CONTRIBUTING.md](CONTRIBUTING.md).
