# MoreOps CLI Command Reference

Справочник всех команд MoreOps CLI.

---

### Общие команды

- `moreops init`
  Интерактивный мастер создания конфигурации `moreops.yaml` с автодетектом языка и фреймворка. Поддерживает выбор модулей одиночно или группами (`1 2 4` или `14`).

- `moreops status`
  Отображение текущего статуса проекта, активных окружений, включенных модулей и наличия сгенерированных манифестов на диске.

- `moreops doctor`
  Диагностика окружения разработчика: проверка наличия `docker`, `k3d`, `kubectl`, `helm`, `tofu`, `ansible`, `vault`, `trivy`, `git`.

- `moreops version`
  Отображение текущей версии MoreOps.

---

### Управление изменениями и манифестами

- `moreops plan [--env <env>] [--config <path>]`
  Генерация манифестов в памяти и расчет unified diff с текущим состоянием на диске без применения мутаций.

- `moreops apply [--env <env>] [--yes]`
  Предпросмотр diff'а изменений и интерактивный запрос на подтверждение (`[y/N]`). Флаг `--yes` отключает интерактивный запрос для автоматизации в CI/CD пайплайнах.

- `moreops generate [--env <env>] [--module <name>]`
  Прямая генерация манифестов конкретного модуля (`monitoring`, `containers`, `kubernetes`, `security`, `storage`, `cicd`) или всей конфигурации.

- `moreops validate [--config <path>] [--security] [--rbac]`
  Синтаксический и семантический аудит конфигурации, а также статическая проверка безопасности манифестов (SEC001-SEC005, REL001, CON001).

---

### Мониторинг и хранилище

- `moreops monitor status`
  Проверка конфигурации Prometheus Operator `ServiceMonitor`, AlertRules и Grafana дашбордов.

- `moreops storage check`
  Валидация настроек S3-провайдера (MinIO, Yandex Cloud, VK Cloud, Selectel, AWS) и расписаний бэкапов Velero.

---

### Безопасность и сканирование

- `moreops scan fs [path]`
  Сканирование файловой системы, зависимостей и манифестов на уязвимости через Trivy (с авто-фоллбэком на Docker, если Trivy не установлен локально).

- `moreops scan image <image_name>`
  Сканирование образа контейнера на уязвимости.

---

### Кластеры и тестирование

- `moreops cluster list`
  Список локальных k3d-кластеров и контекстов kubectl.

- `moreops cluster create <name>`
  Создание изолированного локального k3d Kubernetes кластера.

- `moreops cluster delete <name>`
  Удаление локального k3d кластера.

- `moreops test [--ephemeral]`
  Запуск валидации манифестов. При флаге `--ephemeral` поднимает временный k3d-кластер, запускает dry-run проверку развертывания и автоматически уничтожает кластер после тестов.

---

### Управление окружениями (`moreops env`)

- `moreops env list`
  Отображение списка окружений (`dev`, `staging`, `prod`) с привязанными кластерами и неймспейсами.

- `moreops env switch <name>`
  Переключение активного окружения.

- `moreops env diff <env1> <env2>`
  Сравнение параметров между двумя окружениями.

- `moreops env add <name>`
  Добавление нового окружения в `moreops.yaml`.

---

### Инфраструктура и релизы

- `moreops iac plan`
  Запуск `tofu plan` или `terraform plan` в директории `terraform/`.

- `moreops iac apply [--yes]`
  Применение изменений OpenTofu/Terraform с подтверждением.

- `moreops release [patch|minor|major]`
  Автоматический бамп версии (SemVer), обновление `moreops.yaml`, `Chart.yaml` и запись в `CHANGELOG.md`.
