#!/usr/bin/env bash
set -euo pipefail

CLUSTER_NAME="${1:-moreops-dev}"

echo "🚀 Setting up local k3d cluster: ${CLUSTER_NAME}"

if k3d cluster list | grep -q "${CLUSTER_NAME}"; then
    echo "ℹ️  Cluster ${CLUSTER_NAME} already exists."
else
    k3d cluster create "${CLUSTER_NAME}" \
        --agents 1 \
        --port "8080:80@loadbalancer" \
        --wait

    echo "✅ Cluster ${CLUSTER_NAME} created successfully."
fi

kubectl cluster-info
echo "🎉 Cluster is ready for MoreOps integration testing!"
