#!/usr/bin/env bash
set -euo pipefail

# ==============================================================================
# MoreOps Installer Script
# One-liner installer for MoreOps DevOps CLI.
# ==============================================================================

REPO="${MOREOPS_REPO:-AugurMonne/moreops}"
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

# 1. Check if running inside local source tree with Go
if command -v go >/dev/null 2>&1 && [ -f "go.mod" ]; then
    echo "Found local Go development tree. Building MoreOps directly..."
    CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o "bin/moreops" ./cmd/moreops
    $SUDO cp "bin/moreops" "$INSTALL_DIR/moreops"
    $SUDO chmod +x "$INSTALL_DIR/moreops"
    echo "✓ Successfully installed moreops to $INSTALL_DIR/moreops"
    exit 0
fi

# 2. Check for pre-built binary release on GitHub
if [ "$VERSION" = "latest" ]; then
    RELEASE_URL="https://github.com/${REPO}/releases/latest/download/moreops_${OS}_${ARCH}.tar.gz"
else
    RELEASE_URL="https://github.com/${REPO}/releases/download/${VERSION}/moreops_${OS}_${ARCH}.tar.gz"
fi

TMP_DIR="$(mktemp -d)"
trap 'rm -rf "$TMP_DIR"' EXIT

echo "Checking for pre-built release from GitHub..."
DOWNLOADED=false
if command -v curl >/dev/null 2>&1; then
    if curl -fsSL "$RELEASE_URL" -o "$TMP_DIR/moreops.tar.gz" 2>/dev/null; then
        DOWNLOADED=true
    fi
elif command -v wget >/dev/null 2>&1; then
    if wget -q "$RELEASE_URL" -O "$TMP_DIR/moreops.tar.gz" 2>/dev/null; then
        DOWNLOADED=true
    fi
fi

if [ "$DOWNLOADED" = true ] && [ -f "$TMP_DIR/moreops.tar.gz" ] && [ -s "$TMP_DIR/moreops.tar.gz" ]; then
    tar -xzf "$TMP_DIR/moreops.tar.gz" -C "$TMP_DIR"
    $SUDO cp "$TMP_DIR/moreops" "$INSTALL_DIR/moreops"
    $SUDO chmod +x "$INSTALL_DIR/moreops"
    echo "✓ MoreOps installed successfully to $INSTALL_DIR/moreops"
else
    # 3. Fallback: If no pre-built release on GitHub yet, but user has git and go, clone & build!
    if command -v git >/dev/null 2>&1 && command -v go >/dev/null 2>&1; then
        echo "Pre-built release not yet published on GitHub."
        echo "Building MoreOps from source repository (https://github.com/${REPO})..."
        BUILD_DIR="$(mktemp -d)"
        git clone --depth 1 "https://github.com/${REPO}.git" "$BUILD_DIR"
        (cd "$BUILD_DIR" && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o moreops ./cmd/moreops)
        $SUDO cp "$BUILD_DIR/moreops" "$INSTALL_DIR/moreops"
        $SUDO chmod +x "$INSTALL_DIR/moreops"
        rm -rf "$BUILD_DIR"
        echo "✓ MoreOps compiled and installed successfully to $INSTALL_DIR/moreops"
    else
        echo "❌ Pre-built release not found at: $RELEASE_URL"
        echo "To install from source, please ensure 'git' and 'go' are installed."
        exit 1
    fi
fi

echo ""
echo "Verify installation by running:"
echo "  moreops version"
echo "  moreops doctor"
