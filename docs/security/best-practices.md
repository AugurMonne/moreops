# Security Best Practices in MoreOps

Безопасность по умолчанию (**Secure by Default**) является фундаментальным правилом MoreOps.

---

## 1. Защита контейнеров и подов (Pod Security Standards)

Каждый генерируемый MoreOps под настроен с максимальным уровнем изоляции:

- `runAsNonRoot: true`: запуск процессов строго от непривилегированного пользователя (UID 10001).
- `allowPrivilegeEscalation: false`: блокировка повышения привилегий через SUID бинарники.
- `readOnlyRootFilesystem: true`: файловая система контейнера доступна только для чтения; запись разрешена только во временные тома `emptyDir`.
- `capabilities: { drop: ["ALL"] }`: сброс всех потенциально опасных Linux capabilities.
- `seccompProfile: { type: "RuntimeDefault" }`: стандартный профиль фильтрации системных вызовов.

---

## 2. Изоляция сети (NetworkPolicy)

Для каждого приложения создается манифест `NetworkPolicy`:
- Входящий трафик (`Ingress`) разрешен исключительно на рабочий порт сервиса.
- Исходящий трафик (`Egress`) по умолчанию ограничен системными DNS-запросами (UDP 53) и явными целевыми сервисами.

---

## 3. Управление секретами

- Запрет хранения секретов в открытом виде в Git.
- Использование HashiCorp Vault в качестве единого источника правды.
- Автоматическая синхронизация через External Secrets Operator (ESO) в Kubernetes Secrets.
- Отключение автоматического монтирования токенов ServiceAccount (`automountServiceAccountToken: false`) для предотвращения кражи учетных данных Kubernetes API при компрометации контейнера.

---

## 4. Встроенный аудит (`moreops validate`)

Команда `moreops validate` проверяет соблюдение правил:
- **SEC001**: Контейнер обязан содержать `runAsNonRoot: true`.
- **SEC002**: Контейнер обязан содержать `allowPrivilegeEscalation: false`.
- **SEC003**: Контейнер должен использовать `readOnlyRootFilesystem: true`.
- **SEC004**: Обязательное указание CPU/Memory `requests` и `limits` (защита от DoS ноды).
- **REL001**: Обязательное наличие `livenessProbe` и `readinessProbe`.
- **CON001**: Согласованность селекторов Service с метками Deployment.
