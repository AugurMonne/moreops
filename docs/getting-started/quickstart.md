# Quickstart Guide

MoreOps — универсальный DevOps-оркестратор на Go, управляющий индустриальными инструментами (Docker, Kubernetes, Helm, Vault, S3, Prometheus, Argo CD, OpenTofu, Ansible).

---

## 1. Установка и запуск

Склонируйте репозиторий и соберите бинарник:

```bash
git clone <repository_url>
cd moreops
make build
```

Проверьте окружение:

```bash
./bin/moreops doctor
```

Команда `doctor` выведет список установленных утилит (`docker`, `k3d`, `kubectl`, `helm`, `tofu`, `ansible`, `vault`, `trivy`, `git`).

---

## 2. Инициализация проекта

Запустите интерактивный мастер:

```bash
./bin/moreops init
```

1. Введите имя проекта и версию (нажатие Enter выбирает дефолтное значение `[default]`).
2. Выберите компоненты, введя их номера (например, `1 2 4` или слитно `14`):
   - `[1] Containers & Docker`
   - `[2] Kubernetes Manifests (Helm / Kustomize / Raw)`
   - `[3] Monitoring & Observability`
   - `[4] Security & Secrets (Vault, ESO, RBAC)`
   - `[5] Storage & Backup (MinIO, S3, Velero, DB Backups)`
   - `[6] CI/CD & Automation (GitHub Actions, GitLab CI, ArgoCD)`

После завершения мастер создаст декларативный файл `moreops.yaml`.

---

## 3. Принцип Preview / Diff / Confirm

### Предпросмотр изменений (Dry-Run / Diff):

```bash
./bin/moreops plan --env dev
```

MoreOps построит план в памяти, сравнит его с файлами на диске и выведет цветной unified diff.

### Безопасное применение изменений:

```bash
./bin/moreops apply --env dev
```

Команда покажет preview и потребует явного подтверждения:
`Do you want to perform these actions? (yes/no): `

Для CI/CD автоматизации используйте флаг подтверждения:

```bash
./bin/moreops apply --yes --env prod
```

---

## 4. Валидация и безопасность

Проверьте манифесты на синтаксические ошибки, безопасность и согласованность:

```bash
./bin/moreops validate
```

Запустите проверку безопасности через Trivy:

```bash
./bin/moreops scan fs
```
