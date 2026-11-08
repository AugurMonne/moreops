#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# MoreOps Installer Script
# One-liner installer for MoreOps DevOps CLI.
# ==============================================================================

REPO="forgeops/moreops"
VERSION="${MOREOPS_VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Check sudo requirement
SUDO=""
if [ "$(id -u)" -ne 0 ]; then
    if [ -w "$INSTALL_DIR" ]; then
        SUDO=""
    elif command -v sudo >/dev/null 2>&1; then
        SUDO="sudo"
    else
        INSTALL_DIR="$HOME/.local/bin"
        mkdir -p "$INSTALL_DIR"
    fi
fi

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$ARCH" in
    x86_64|amd64)
        ARCH="amd64"
        ;;
    arm64|aarch64)
        ARCH="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

echo "=========================================="
echo "      Installing MoreOps CLI ($OS/$ARCH)  "
echo "=========================================="

# Check if Go is installed to build from source if release archive is unavailable
if command -v go >/dev/null 2>&1 && [ -f "go.mod" ]; then
    echo "Found local Go development tree. Building MoreOps directly..."
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "bin/moreops" ./cmd/moreops
    $SUDO cp "bin/moreops" "$INSTALL_DIR/moreops"
    $SUDO chmod +x "$INSTALL_DIR/moreops"
    echo "✓ Successfully installed moreops to $INSTALL_DIR/moreops"
    exit 0
fi

# Fallback download logic from GitHub Releases
if [ "$VERSION" = "latest" ]; then
    RELEASE_URL="https://github.com/${REPO}/releases/latest/download/moreops_${OS}_${ARCH}.tar.gz"
else
    RELEASE_URL="https://github.com/${REPO}/releases/download/${VERSION}/moreops_${OS}_${ARCH}.tar.gz"
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

echo "Downloading MoreOps from GitHub..."
if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$RELEASE_URL" -o "$TMP_DIR/moreops.tar.gz" || true
elif command -v wget >/dev/null 2>&1; then
    wget -q "$RELEASE_URL" -O "$TMP_DIR/moreops.tar.gz" || true
fi

if [ -f "$TMP_DIR/moreops.tar.gz" ] && [ -s "$TMP_DIR/moreops.tar.gz" ]; then
    tar -xzf "$TMP_DIR/moreops.tar.gz" -C "$TMP_DIR"
    $SUDO cp "$TMP_DIR/moreops" "$INSTALL_DIR/moreops"
    $SUDO chmod +x "$INSTALL_DIR/moreops"
    echo "✓ MoreOps installed successfully to $INSTALL_DIR/moreops"
else
    # If downloading prebuilt binary fails and go is installed, try go install
    if command -v go >/dev/null 2>&1; then
        echo "Pre-built release not yet available. Installing via 'go install'..."
        go install "github.com/${REPO}/cmd/moreops@latest"
        echo "✓ Installed via go install ($GOPATH/bin/moreops or ~/go/bin/moreops)"
    else
        echo "❌ Installation failed. Please install Go or download pre-compiled binary manually."
        exit 1
    fi
fi

echo ""
echo "Verify installation by running:"
echo "  moreops version"
echo "  moreops doctor"
