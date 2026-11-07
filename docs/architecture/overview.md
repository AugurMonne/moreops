# MoreOps Architecture & Principles

MoreOps построен на ключевом принципе: **Оркестрация существующих проверенных инструментов, а не изобретение собственных аналогов.**

---

## Ключевые принципы

1. **Оркестрация индустриальных стандартов**:
   MoreOps управляет Helm, Kustomize, HashiCorp Vault, External Secrets Operator, Prometheus Operator, Velero, OpenTofu, Ansible и Argo CD.

2. **Модульность и селективность**:
   Каждый модуль независим (`internal/modules/*`). Пользователь выбирает только то, что требуется проекту.

3. **Гарантия Preview / Diff / Confirm**:
   Ни одно изменение инфраструктуры или файлов манифестов не вносится вслепую. Движок diff (`internal/core/diff`) вычисляет unified diff перед применением.

4. **Multi-Cluster & Multi-Environment**:
   Поддержка нескольких сред (`dev`, `staging`, `prod`) с привязкой к разным Kubernetes контекстам/кластерам, неймспейсам и облакам.

5. **Безопасность по умолчанию (Secure by default)**:
   - Все генерируемые поды имеют `runAsNonRoot: true`, `allowPrivilegeEscalation: false`, `readOnlyRootFilesystem: true`, `drop: [ALL]`.
   - Выделенные ServiceAccount с `automountServiceAccountToken: false`.
   - Встроенный Security Linter в `moreops validate`.
   - Сканирование уязвимостей через Trivy.

6. **Тестирование на временных кластерах**:
   Интеграция с `k3d` для запуска изолированных временных Kubernetes-кластеров на время тестов (`moreops test --ephemeral`).

---

## Архитектурная схема компонентов

```
                  +---------------------------+
                  |        MoreOps CLI        |
                  | (Cobra Commands & Wizard) |
                  +-------------+-------------+
                                |
             +------------------+------------------+
             |                                     |
+------------v------------+              +---------v---------+
|    Core Config Engine   |              |  Workflow Engine  |
|  (moreops.yaml & Envs)  |              |  (Diff & Confirm) |
+------------+------------+              +---------+---------+
             |                                     |
             +------------------+------------------+
                                |
                   +------------v------------+
                   |  Plugin/Module Registry |
                   +------------+------------+
                                |
    +------------+------------+--+----------+------------+------------+
    |            |            |             |            |            |
+---v----+  +----v-----+ +----v----+   +----v-----+ +----v----+  +----v----+
| Docker |  |   K8s    | | Security|   |Monitoring| | Storage |  |  CI/CD  |
| Compose|  |Helm/Kust | |Vault/ESO|   |Prom/Graf | |S3/Velero|  |Jenkins/ |
|        |  |          | |         |   |          | |         |  |ArgoCD   |
+--------+  +----------+ +---------+   +----------+ +---------+  +---------+
```
