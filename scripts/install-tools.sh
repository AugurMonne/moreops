#!/usr/bin/env bash
set -euo pipefail

echo "=========================================="
echo " MoreOps CLI & Dependencies Installer"
echo "=========================================="

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check docker
if ! command_exists docker; then
    echo "⚠️  Docker is not installed. Please install Docker to use container and k3d features."
fi

# Check kubectl
if ! command_exists kubectl; then
    echo "📦 Installing kubectl..."
    curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
    chmod +x kubectl
    sudo mv kubectl /usr/local/bin/
else
    echo "✅ kubectl is installed: $(kubectl version --client -o json 2>/dev/null | grep gitVersion || true)"
fi

# Check helm
if ! command_exists helm; then
    echo "📦 Installing helm..."
    curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
else
    echo "✅ helm is installed: $(helm version --short)"
fi

# Check k3d
if ! command_exists k3d; then
    echo "📦 Installing k3d..."
    curl -s https://raw.githubusercontent.com/k3d-io/k3d/main/install.sh | bash
else
    echo "✅ k3d is installed: $(k3d version | head -n 1)"
fi

echo "=========================================="
echo " Build & install MoreOps binary"
echo "=========================================="
go build -ldflags "-s -w" -o bin/moreops cmd/moreops/main.go
echo "✅ MoreOps installed at $(pwd)/bin/moreops"
