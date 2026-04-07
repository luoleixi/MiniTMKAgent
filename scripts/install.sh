#!/bin/bash
# MiniTMK Agent - One-Click Installer for macOS/Linux
# Usage: curl -fsSL https://raw.githubusercontent.com/luoleixi/MiniTMKAgent/main/scripts/install.sh | bash

set -e

# 配置
REPO_OWNER="luoleixi"
REPO_NAME="MiniTMKAgent"
BINARY_NAME="mini-tmk-agent"
VERSION="${MINI_TMK_VERSION:-latest}"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

info() { echo -e "${CYAN}[INFO]${NC} $1"; }
success() { echo -e "${GREEN}[OK]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; }

# 检测操作系统
detect_os() {
    OS=$(uname -s | tr '[:upper:]' '[:lower:]')
    case "$OS" in
        linux*) echo "linux" ;;
        darwin*) echo "darwin" ;;
        *) error "不支持的操作系统: $OS"; exit 1 ;;
    esac
}

# 检测架构
detect_arch() {
    ARCH=$(uname -m)
    case "$ARCH" in
        x86_64) echo "amd64" ;;
        amd64) echo "amd64" ;;
        arm64) echo "arm64" ;;
        aarch64) echo "arm64" ;;
        *) error "不支持的架构: $ARCH"; exit 1 ;;
    esac
}

# 获取最新版本
get_latest_version() {
    info "获取最新版本..."
    local api_url="https://api.github.com/repos/${REPO_OWNER}/${REPO_NAME}/releases/latest"
    VERSION=$(curl -fsSL --connect-timeout 10 "$api_url" 2>/dev/null | grep -o '"tag_name": "[^"]*"' | cut -d'"' -f4)
    if [ -z "$VERSION" ]; then
        error "无法获取最新版本"
        exit 1
    fi
    info "最新版本: $VERSION"
}

# 下载二进制文件
download_binary() {
    local os=$1
    local arch=$2
    local asset_name="${BINARY_NAME}-${os}-${arch}"
    local download_url="https://github.com/${REPO_OWNER}/${REPO_NAME}/releases/download/${VERSION}/${asset_name}"
    local output_path="/tmp/${BINARY_NAME}"

    info "下载 ${asset_name}..."

    if ! curl -fsSL --connect-timeout 30 --max-time 120 "$download_url" -o "$output_path"; then
        error "下载失败: $download_url"
        exit 1
    fi

    # 检查文件大小
    local size=$(stat -f%z "$output_path" 2>/dev/null || stat -c%s "$output_path" 2>/dev/null)
    if [ "$size" -lt 1000 ]; then
        error "下载文件无效"
        exit 1
    fi

    local size_mb=$(echo "scale=2; $size / 1024 / 1024" | bc 2>/dev/null || echo "$((size / 1024 / 1024))")
    success "下载完成 (${size_mb} MB)"
    echo "$output_path"
}

# 安装二进制文件
install_binary() {
    local source_path=$1
    local install_dir=""

    # 优先安装到 /usr/local/bin，如果没有权限则安装到 ~/.local/bin
    if [ -w "/usr/local/bin" ]; then
        install_dir="/usr/local/bin"
    elif [ -d "$HOME/.local/bin" ] && [ -w "$HOME/.local/bin" ]; then
        install_dir="$HOME/.local/bin"
    else
        # 创建 ~/.local/bin
        mkdir -p "$HOME/.local/bin"
        install_dir="$HOME/.local/bin"
    fi

    local target_path="${install_dir}/${BINARY_NAME}"

    info "安装到: $target_path"

    # 移动并设置权限
    mv -f "$source_path" "$target_path"
    chmod +x "$target_path"

    success "安装完成"
    echo "$target_path"
}

# 添加到 PATH
add_to_path() {
    local install_dir=$1

    # 检查是否已在 PATH 中
    if echo "$PATH" | grep -q "$install_dir"; then
        info "PATH 中已存在"
        return
    fi

    # 确定 shell 配置文件
    local shell_rc=""
    if [ -n "$ZSH_VERSION" ]; then
        shell_rc="$HOME/.zshrc"
    elif [ -n "$BASH_VERSION" ]; then
        shell_rc="$HOME/.bashrc"
    else
        shell_rc="$HOME/.profile"
    fi

    info "添加到 PATH ($shell_rc)..."
    echo "export PATH=\"$install_dir:\$PATH\"" >> "$shell_rc"
    success "已添加到 PATH"
    warn "请运行: source $shell_rc 或重新打开终端以使更改生效"
}

# 主程序
echo ""
echo "========================================"
echo "  MiniTMK Agent 一键安装"
echo "========================================"
echo ""

OS=$(detect_os)
ARCH=$(detect_arch)
info "系统: $OS/$ARCH"

if [ "$VERSION" = "latest" ]; then
    get_latest_version
fi
info "安装版本: $VERSION"

TEMP_FILE=$(download_binary "$OS" "$ARCH")
INSTALLED_PATH=$(install_binary "$TEMP_FILE")

# 获取安装目录
INSTALL_DIR=$(dirname "$INSTALLED_PATH")
add_to_path "$INSTALL_DIR"

echo ""
echo "========================================"
echo "  安装完成!"
echo "========================================"
echo ""
echo "使用方法:"
echo ""
echo "  $BINARY_NAME quickstart"
echo ""
echo "或者完整路径:"
echo "  $INSTALLED_PATH quickstart"
echo ""
echo "获取 API Key: https://dashscope.console.aliyun.com/"
echo ""
