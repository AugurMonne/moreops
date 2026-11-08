# MoreOps CLI Command Reference

Справочник всех команд MoreOps CLI.

---

### Общие команды

- `moreops init`
  Интерактивный мастер создания конфигурации `moreops.yaml` с автодетектом языка и фреймворка. Поддерживает полноценную TUI-навигацию стрелками (`↑`/`↓` или `j`/`k`), выбор клавишей `Space`, выбор всех по `a`, мгновенный переход по цифрам `1-N`, а также безопасный fallback на ввод чисел для не-TTY окружений и CI/CD.

- `moreops status [--live]`
  Отображение текущего статуса проекта, активных окружений, включенных модулей (Containers, Kubernetes, Database, Messaging, Monitoring, Security, Storage, CI/CD). Флаг `--live` выполняет инспекцию состояния реальных подов в активном кластере Kubernetes.

- `moreops doctor [--fix]`
  Комплексная диагностика окружения разработчика: проверка Docker daemon, kubectl, helm, git, прав доступа к файлам (0600) и аудит утечек токенов в конфигах. Флаг `--fix` (`-f`) автоматически исправляет небезопасные права файлов и добавляет скрытые секреты в `.gitignore`.

- `moreops diagnose [--env <env>] [--namespace <ns>]` (алиасы: `troubleshoot`, `debug`)
  Интеллектуальная диагностика проблемных рабочих нагрузок Kubernetes: детектирование `CrashLoopBackOff`, `OOMKilled` (Exit Code 137), `ImagePullBackOff`, циклических перезапусков и последних Warning-событий с выдачей готовых рекомендаций по устранению сбоев.

- `moreops version`
  Отображение текущей версии MoreOps.

- `moreops completion [bash|zsh|fish]`
  Генерация скрипта автодополнения для вашей командной оболочки.

---

### Управление изменениями и манифестами

- `moreops plan [--env <env>] [--config <path>] [--summary]`
  Генерация манифестов в памяти и расчет unified diff с текущим состоянием на диске без применения мутаций. Флаг `--summary` (`-s`) выводит компактную таблицу изменений по модулям без сырого текста diff'а.

- `moreops apply [--env <env>] [--yes]`
  Предпросмотр diff'а изменений, интерактивный запрос на подтверждение (`[y/N]`) и пошаговая синхронизация с компактной индикацией статуса каждого активного модуля (`Containers`, `Kubernetes`, `Database`, `Messaging`, `Monitoring`, `CI/CD`). Флаг `--yes` отключает интерактивный запрос для автоматизации в CI/CD пайплайнах.

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
