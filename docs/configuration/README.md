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
