#!/usr/bin/env bash
set -euo pipefail

ROLE_NAME="${1:-app-role}"
SERVICE_ACCOUNT="${2:-app-sa}"
NAMESPACE="${3:-default}"

echo "Configuring Vault Kubernetes Auth for role: ${ROLE_NAME}"

vault auth enable kubernetes || true

vault write auth/kubernetes/config \
    kubernetes_host="https://kubernetes.default.svc.cluster.local:443"

vault write auth/kubernetes/role/"${ROLE_NAME}" \
    bound_service_account_names="${SERVICE_ACCOUNT}" \
    bound_service_account_namespaces="${NAMESPACE}" \
    policies="${ROLE_NAME}" \
    ttl=24h

echo "✅ Kubernetes Auth configured successfully for ${ROLE_NAME}"
