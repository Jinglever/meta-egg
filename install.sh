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

# 安装 shell completion
install_completion() {
  SHELL_NAME=$(basename "$SHELL")
  case "$SHELL_NAME" in
    zsh)
      ZSH_FPATHS=$(zsh -c 'print -l $fpath')
      if echo "$ZSH_FPATHS" | grep -qx '/usr/local/share/zsh/site-functions'; then
        COMPLETION_DIR="/usr/local/share/zsh/site-functions"
        if sudo test -w "$COMPLETION_DIR" 2>/dev/null; then
          sudo "$INSTALL_DIR/$BINARY_NAME" completion zsh | sudo tee "$COMPLETION_DIR/_meta-egg" > /dev/null
          echo "Zsh completion installed to $COMPLETION_DIR/_meta-egg"
        else
          "$INSTALL_DIR/$BINARY_NAME" completion zsh > "$HOME/.meta-egg-completion.zsh"
          echo "No write permission to $COMPLETION_DIR, installed to $HOME/.meta-egg-completion.zsh"
          echo "Add 'source ~/.meta-egg-completion.zsh' to your ~/.zshrc to enable completion."
        fi
      elif echo "$ZSH_FPATHS" | grep -qx '/usr/share/zsh/site-functions'; then
        COMPLETION_DIR="/usr/share/zsh/site-functions"
        if sudo test -w "$COMPLETION_DIR" 2>/dev/null; then
          sudo "$INSTALL_DIR/$BINARY_NAME" completion zsh | sudo tee "$COMPLETION_DIR/_meta-egg" > /dev/null
          echo "Zsh completion installed to $COMPLETION_DIR/_meta-egg"
        else
          "$INSTALL_DIR/$BINARY_NAME" completion zsh > "$HOME/.meta-egg-completion.zsh"
          echo "No write permission to $COMPLETION_DIR, installed to $HOME/.meta-egg-completion.zsh"
          echo "Add 'source ~/.meta-egg-completion.zsh' to your ~/.zshrc to enable completion."
        fi
      else
        "$INSTALL_DIR/$BINARY_NAME" completion zsh > "$HOME/.meta-egg-completion.zsh"
        echo "Zsh completion installed to $HOME/.meta-egg-completion.zsh"
        echo "Add 'source ~/.meta-egg-completion.zsh' to your ~/.zshrc to enable completion."
      fi
      ;;
    bash)
      # 优先用 /etc/bash_completion.d，若无权限则用 ~/.meta-egg-completion.bash
      if sudo test -w /etc/bash_completion.d 2>/dev/null; then
        COMPLETION_DIR="/etc/bash_completion.d"
        sudo "$INSTALL_DIR/$BINARY_NAME" completion bash | sudo tee "$COMPLETION_DIR/meta-egg" > /dev/null
        echo "Bash completion installed to $COMPLETION_DIR/meta-egg"
      else
        COMPLETION_FILE="$HOME/.meta-egg-completion.bash"
        "$INSTALL_DIR/$BINARY_NAME" completion bash > "$COMPLETION_FILE"
        echo "Bash completion installed to $COMPLETION_FILE"
        echo "Add 'source $COMPLETION_FILE' to your ~/.bashrc to enable completion."
      fi
      ;;
    fish)
      COMPLETION_DIR="$HOME/.config/fish/completions"
      mkdir -p "$COMPLETION_DIR"
      "$INSTALL_DIR/$BINARY_NAME" completion fish > "$COMPLETION_DIR/meta-egg.fish"
      echo "Fish completion installed to $COMPLETION_DIR/meta-egg.fish"
      ;;
    *)
      echo "Unknown shell: $SHELL_NAME. You can generate completion manually with 'meta-egg completion <shell>'"
      ;;
  esac
}

install_completion

echo "If completion does not work immediately, try restarting your terminal." 