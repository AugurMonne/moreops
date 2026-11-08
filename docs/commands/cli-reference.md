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

- `moreops completion [bash|zsh|fish]`
  Генерация скрипта автодополнения для вашей командной оболочки.

---

### Управление изменениями и манифестами

- `moreops plan [--env <env>] [--config <path>]`
  Генерация манифестов в памяти и расчет unified diff с текущим состоянием на диске без применения мутаций.

- `moreops apply [--env <env>] [--yes]`
  Предпросмотр diff'а изменений и интерактивный запрос на подтверждение (`[y/N]`). Флаг `--yes` отключает интерактивный запрос для автоматизации в CI/CD пайплайнах.

- `moreops generate [--env <env>] [--module <name>]`
  Прямая генерация манифестов конкретного модуля (`monitoring`, `containers`, `kubernetes`, `security`, `storage`, `cicd`) или всей конфигурации.

- `moreops validate [--config <path>] [--security] [--rbac]`
  Синтаксический и семантический аудит конфигурации, а также статическая проверка безопасности манифестов (SEC001-SEC009, REL001, CON001).

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

### Helm OCI и Релизы

- `moreops chart package [path]`
  Упаковка Helm-чарта приложения в архив `.tgz`.

- `moreops chart push <archive.tgz> <oci://registry/path>`
  Публикация упакованного чарта в корпоративный OCI-реджистри (GitHub Packages, Yandex Container Registry, Harbor).

- `moreops release [patch|minor|major]`
  Автоматический бамп версии (SemVer), обновление `moreops.yaml`, `Chart.yaml` и запись в `CHANGELOG.md`.

---

### Инфраструктура (IaC)

- `moreops iac plan`
  Запуск `tofu plan` или `terraform plan` в директории `terraform/`.

- `moreops iac apply [--yes]`
  Применение изменений OpenTofu/Terraform с подтверждением.

---

### Кастомизация шаблонов и защита файлов (`moreops template`)

- `moreops template list`
  Вывод каталога доступных для кастомизации шаблонов (Dockerfile, Helm deployment, values, GitLab CI, GitHub Actions, Jenkins).

- `moreops template eject [template-id | all] [--dir .moreops/templates]`
  Извлечение встроенных шаблонов в `.moreops/templates/` для свободного редактирования под нужды open-source проекта. MoreOps автоматически использует файлы из этой директории при генерации.

- **Защита ручных правок:** Добавление комментария `# moreops:ignore` или `# moreops:custom` в первые 10 строк любого сгенерированного файла защищает его от перезаписи командами `moreops apply` и `moreops plan`.

---

### Управление секретами (`moreops secrets`)

- `moreops secrets init`
  Инициализация локального ключа шифрования AES-256-GCM (`.moreops/secrets.key` с правами `0600`) и конфигурации `.sops.yaml`. Ключ автоматически добавляется в `.gitignore`.

- `moreops secrets template [name] [--backend vault|aws|gcp|azure] [--namespace default]`
  Генерация готовых манифестов `SecretStore` и `ExternalSecret` для External Secrets Operator.

- `moreops secrets encrypt <file>`
  Шифрование файла с секретами через SOPS или встроенный нативный AES-256-GCM.

- `moreops secrets decrypt <file>`
  Расшифровка секретов для редактирования или деплоя.

---

### Диагностика и статус (`moreops doctor`, `moreops status`)

- `moreops doctor`
  Комплексный аудит инструментов разработчика (Docker daemon, kubectl, helm, git), валидности конфигураций, отсутствия открытых токенов/паролей в кодовой базе и проверки прав доступа к чувствительным файлам.

- `moreops status [--live] [--env dev]`
  Отображение состояния проекта и активных модулей. Флаг `--live` выполняет опрос живого Kubernetes-кластера, проверяя статус подов сервиса, HPA и доступность неймспейса.
