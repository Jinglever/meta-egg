#!/usr/bin/env bash

set -e

REPO="Jinglever/meta-egg"
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="meta-egg"

# 检测架构
ARCH=$(uname -m)
case "$ARCH" in
    x86_64|amd64) ARCH="amd64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

# 检测操作系统
OS=$(uname | tr '[:upper:]' '[:lower:]')
if [[ "$OS" != "linux" ]]; then
    echo "This script only supports Linux."
    exit 1
fi

# 检查 unzip 是否安装
if ! command -v unzip >/dev/null 2>&1; then
    echo "unzip is required but not installed. Please install unzip and try again."
    exit 1
fi

# 获取最新版本号
LATEST=$(curl -sL "https://api.github.com/repos/$REPO/releases/latest" | grep tag_name | cut -d '"' -f4)
if [[ -z "$LATEST" ]]; then
    echo "Failed to fetch latest version."
    exit 1
fi

# 组装下载链接
ZIPFILE="meta-egg-${OS}-${ARCH}.zip"
URL="https://github.com/$REPO/releases/download/$LATEST/$ZIPFILE"

# 下载并解压
TMP_DIR=$(mktemp -d)
cd "$TMP_DIR"
echo "Downloading $URL ..."
curl -LO "$URL"

echo "Extracting $ZIPFILE ..."
unzip "$ZIPFILE"

echo "Installing $BINARY_NAME to $INSTALL_DIR ..."
sudo mv "$BINARY_NAME" "$INSTALL_DIR/"
sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"

cd /
rm -rf "$TMP_DIR"

echo "meta-egg ($LATEST) installed successfully!"
echo "Run 'meta-egg --help' to get started." 