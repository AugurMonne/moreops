# Воркфлоу MoreOps (Workflows)

В каталоге `workflows/` содержатся готовые сценарии автоматизации для платформы:

- `workflows/application/deploy.yaml`: Полный цикл сборки, статического аудита, эфемерного тестирования и релиза через GitOps.
- `workflows/monitoring/setup-observability.yaml`: Развертывание ServiceMonitor, AlertRules и дашбордов Grafana.
- `workflows/backup/disaster-recovery.yaml`: Настройка расписаний создания снимков Velero и потоковой выгрузки дампов баз данных в S3.
- `workflows/security/audit.yaml`: Сканирование уязвимостей через Trivy и проверка прав доступа RBAC.
- `workflows/infrastructure/cluster-provision.yaml`: Поднятие и удаление эфемерных `k3d` кластеров для тестирования.
