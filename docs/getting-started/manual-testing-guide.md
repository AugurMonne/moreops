# Руководство по ручной проверке MoreOps (Manual Testing Guide)

Пошаговый сценарий для проверки всех возможностей MoreOps в терминале от сборки до релиза.

---

## Шаг 0: Сборка и проверка окружения

1. Перейдите в корень репозитория MoreOps и соберите бинарник:
   ```bash
   make build
   ```
2. Проверьте диагностику инструментов:
   ```bash
   ./bin/moreops doctor
   ```
   *Ожидаемый результат*: Таблица доступности утилит (`docker`, `kubectl`, `helm`, `git` и т.д.).

---

## Шаг 1: Создание тестового приложения и `moreops init`

1. Создайте временную папку для тестирования (например, простой Go-сервис):
   ```bash
   mkdir -p /tmp/demo-service && cd /tmp/demo-service
   cat << 'EOF' > main.go
   package main
   import "net/http"
   func main() {
       http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("ok")) })
       http.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("# HELP test\n")) })
       http.ListenAndServe(":8080", nil)
   }
   EOF
   go mod init demo-service
   ```

2. Запустите интерактивный мастер MoreOps:
   ```bash
   /home/augur/Documents/forgeops/bin/moreops init
   ```
   - MoreOps автоматически определит: `🔍 Detected: GO`.
   - Нажмите `Enter` на всех запросах, чтобы принять дефолты (или выберите компоненты, например `1 2 3 4 5 6` или `14`).
   *Ожидаемый результат*: Создан файл `moreops.yaml`.

---

## Шаг 2: Предпросмотр изменений (`moreops plan`)

Запустите расчет плана и unified diff:
```bash
/home/augur/Documents/forgeops/bin/moreops plan
```
*Ожидаемый результат*:
- В терминале отобразится цветной diff:
  - `+++ Dockerfile (new file)`
  - `+++ .dockerignore (new file)`
  - `+++ docker-compose.yml (new file)`
  - `+++ charts/demo-service/...`
  - `+++ Jenkinsfile (new file)` (если выбран Jenkins) или `.github/...`
- Строка: `Plan summary: X to create, 0 to update, 0 to delete, 0 unchanged.`
- **Важно**: ни один файл на диске пока не создан!

---

## Шаг 3: Применение изменений (`moreops apply`)

1. Примените изменения с интерактивным подтверждением:
   ```bash
   /home/augur/Documents/forgeops/bin/moreops apply
   ```
   - Введите `yes` в ответ на запрос `Do you want to perform these actions? (yes/no):`.
   *Ожидаемый результат*:
   - `✓ Apply completed successfully!`
   - На диске появились все файлы: `Dockerfile`, `charts/demo-service/`, `Jenkinsfile`, и т.д.

2. Повторно выполните `plan`:
   ```bash
   /home/augur/Documents/forgeops/bin/moreops plan
   ```
   *Ожидаемый результат*: `No infrastructure changes planned.`

---

## Шаг 4: Проверка статуса (`moreops status`)

Выполните команду проверки состояния:
```bash
/home/augur/Documents/forgeops/bin/moreops status
```
*Ожидаемый результат*:
- Вывод имени проекта, версии, активных окружений (`dev`).
- Список подключенных модулей (`Containers`, `Kubernetes`, `Monitoring`, `Security`, `Storage`, `CI/CD`).
- Список фактически существующих на диске манифестов.

---

## Шаг 5: Аудит безопасности и статический анализ (`moreops validate`)

1. Запустите полную проверку манифестов:
   ```bash
   /home/augur/Documents/forgeops/bin/moreops validate
   ```
   *Ожидаемый результат*:
   - `✓ Configuration syntax and parameters valid.`
   - `🛡 Running security & consistency audit on charts/...`
   - `✓ Manifest audit passed: No security or consistency issues found!`

2. Проверьте фильтрацию аудита:
   ```bash
   /home/augur/Documents/forgeops/bin/moreops validate --security
   /home/augur/Documents/forgeops/bin/moreops validate --rbac
   ```

3. **Проверка срабатывания линтера (тест на ошибку)**:
   Добавьте в любой манифест небезопасную конструкцию (например, `privileged: true` или `hostNetwork: true`):
   ```bash
   echo "      hostNetwork: true" >> charts/demo-service/templates/deployment.yaml
   /home/augur/Documents/forgeops/bin/moreops validate
   ```
   *Ожидаемый результат*:
   - Линтер выведет `❌ [SEC006] Deployment/...: Pod spec must not set hostNetwork: true` и завершится с ошибкой.
   - Удалите тестовую строку после проверки:
   ```bash
   sed -i '/hostNetwork: true/d' charts/demo-service/templates/deployment.yaml
   ```

---

## Шаг 6: Проверка мониторинга и хранилища

1. Проверка конфигурации мониторинга:
   ```bash
   /home/augur/Documents/forgeops/bin/moreops monitor status
   ```
2. Проверка настроек S3 / Velero:
   ```bash
   /home/augur/Documents/forgeops/bin/moreops storage check
   ```

---

## Шаг 7: Управление средами (`moreops env`)

1. Просмотр текущих сред:
   ```bash
   /home/augur/Documents/forgeops/bin/moreops env list
   ```
2. Добавление новой среды (например, `staging`):
   ```bash
   /home/augur/Documents/forgeops/bin/moreops env add staging --cluster k8s-staging --namespace staging
   ```
3. Сравнение параметров сред:
   ```bash
   /home/augur/Documents/forgeops/bin/moreops env diff dev staging
   ```
4. Переключение среды:
   ```bash
   /home/augur/Documents/forgeops/bin/moreops env switch staging
   ```

---

## Шаг 8: Упаковка Helm-чарта (`moreops chart`)

1. Упакуйте созданный Helm-чарт:
   ```bash
   /home/augur/Documents/forgeops/bin/moreops chart package charts/demo-service
   ```
   *Ожидаемый результат*: Создан архив `demo-service-0.1.0.tgz`.

2. Для публикации в OCI (при наличии доступа к реестру):
   ```bash
   # /home/augur/Documents/forgeops/bin/moreops chart push demo-service-0.1.0.tgz oci://cr.yandex/crp12345/charts
   ```

---

## Шаг 9: Автоматизация версий и релизов (`moreops release`)

1. Выполните бамп версии:
   ```bash
   /home/augur/Documents/forgeops/bin/moreops release minor
   ```
   *Ожидаемый результат*:
   - `Bumping version: 0.1.0 -> 0.2.0 (minor)`
   - `✓ Updated moreops.yaml`
   - `✓ Updated charts/demo-service/Chart.yaml to version 0.2.0`
   - `✓ Updated CHANGELOG.md with release notes for v0.2.0`

2. Проверьте `CHANGELOG.md`:
   ```bash
   cat CHANGELOG.md
   ```

---

## Шаг 10: Очистка тестовых данных

После завершения ручной проверки удалите тестовую папку:
```bash
rm -rf /tmp/demo-service
```
