#!/bin/bash

set -e

INSTALLER_VERSION="1.0.0"
REPO="zhaiyz/ocsession"
BINARY_NAME="ocsession"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

info()    { echo -e "${BLUE}ℹ${NC} $1"; }
success() { echo -e "${GREEN}✓${NC} $1"; }
warn()    { echo -e "${YELLOW}!${NC} $1"; }
error()   { echo -e "${RED}✗${NC} $1"; }

detect_os() {
    case "$(uname -s)" in
        Darwin*) echo "macos" ;;
        Linux*)  echo "linux" ;;
        *)       error "不支持的操作系统: $(uname -s)"; exit 1 ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        arm64|aarch64) echo "arm64" ;;
        x86_64|amd64)  echo "amd64" ;;
        *)             error "不支持的架构: $(uname -m)"; exit 1 ;;
    esac
}

get_install_dir() {
    local custom_dir="${INSTALL_DIR:-}"
    
    if [ -n "$custom_dir" ]; then
        mkdir -p "$custom_dir" 2>/dev/null || true
        if [ -w "$custom_dir" ]; then
            echo "$custom_dir"
            return
        fi
        error "自定义路径无写权限: $custom_dir"
        exit 1
    fi
    
    local local_bin="$HOME/.local/bin"
    mkdir -p "$local_bin" 2>/dev/null || true
    if [ -w "$local_bin" ]; then
        echo "$local_bin"
        return
    fi
    
    if [ -d "$local_bin" ]; then
        chmod u+w "$local_bin" 2>/dev/null || true
        if [ -w "$local_bin" ]; then
            echo "$local_bin"
            return
        fi
    fi
    
    local home_bin="$HOME/bin"
    mkdir -p "$home_bin" 2>/dev/null || true
    if [ -w "$home_bin" ]; then
        echo "$home_bin"
        return
    fi
    
    warn "无法创建用户目录，将安装到当前目录"
    echo "$(pwd)"
}

check_path() {
    local dir="$1"
    case ":$PATH:" in
        *":$dir:"*) return 0 ;;
    esac
    return 1
}

suggest_add_to_path() {
    local dir="$1"
    local shell_config=""
    
    case "$SHELL" in
        */zsh)   shell_config="$HOME/.zshrc" ;;
        */bash)  
            if [ -f "$HOME/.bashrc" ]; then
                shell_config="$HOME/.bashrc"
            elif [ -f "$HOME/.bash_profile" ]; then
                shell_config="$HOME/.bash_profile"
            fi
            ;;
        *)       shell_config="$HOME/.profile" ;;
    esac
    
    echo ""
    warn "目录 $dir 不在 PATH 中"
    echo ""
    echo -e "${BOLD}请将以下内容添加到 $shell_config:${NC}"
    echo ""
    echo "    export PATH=\"$dir:\$PATH\""
    echo ""
    echo -e "${BOLD}然后运行:${NC}"
    echo ""
    echo "    source $shell_config"
    echo ""
}

download_file() {
    local url="$1"
    local output="$2"
    
    if command -v curl &> /dev/null; then
        curl -fsSL --connect-timeout 15 "$url" -o "$output"
    elif command -v wget &> /dev/null; then
        wget -q --timeout=15 "$url" -O "$output"
    else
        error "需要 curl 或 wget"
        exit 1
    fi
}

main() {
    echo ""
    echo -e "${BLUE}╔══════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║   OpenCode Session Manager Installer ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════╝${NC}"
    echo ""
    
    local os=$(detect_os)
    local arch=$(detect_arch)
    info "检测到平台: $os-$arch"
    
    local install_dir=$(get_install_dir)
    info "安装路径: $install_dir"
    
    local version="${VERSION:-latest}"
    local filename="$BINARY_NAME-$os-$arch.tar.gz"
    local sha_filename="$BINARY_NAME-$os-$arch.sha256"
    
    if [ "$version" = "latest" ]; then
        local tar_url="https://github.com/$REPO/releases/latest/download/$filename"
        local sha_url="https://github.com/$REPO/releases/latest/download/$sha_filename"
    else
        local tar_url="https://github.com/$REPO/releases/download/$version/$filename"
        local sha_url="https://github.com/$REPO/releases/download/$version/$sha_filename"
    fi
    
    info "下载: $tar_url"
    
    local tmp_dir=$(mktemp -d)
    trap "rm -rf $tmp_dir" EXIT
    
    local tar_file="$tmp_dir/$filename"
    local sha_file="$tmp_dir/$sha_filename"
    
    if ! download_file "$tar_url" "$tar_file"; then
        error "下载失败"
        echo ""
        echo "请检查:"
        echo "  1. 网络连接是否正常"
        echo "  2. GitHub 是否可访问"
        echo "  3. 平台是否支持: $os-$arch"
        echo ""
        echo "手动下载: https://github.com/$REPO/releases"
        exit 1
    fi
    
    info "下载校验文件..."
    if download_file "$sha_url" "$sha_file"; then
        info "验证 SHA256..."
        if command -v shasum &> /dev/null; then
            if ! shasum -a 256 -c "$sha_file" &> /dev/null; then
                error "校验失败，文件可能已损坏"
                exit 1
            fi
            success "校验通过"
        else
            warn "shasum 不可用，跳过验证"
        fi
    else
        warn "无法下载校验文件，跳过验证"
    fi
    
    info "解压..."
    tar -xzf "$tar_file" -C "$tmp_dir"
    
    info "安装..."
    local binary_path="$install_dir/$BINARY_NAME"
    
    if [ -f "$binary_path" ]; then
        local backup_path="$binary_path.backup"
        mv "$binary_path" "$backup_path"
        info "备份旧版本到: $backup_path"
    fi
    
    mv "$tmp_dir/$BINARY_NAME" "$binary_path"
    chmod +x "$binary_path"
    
    if ! [ -x "$binary_path" ]; then
        error "安装失败"
        exit 1
    fi
    
    success "安装成功!"
    
    echo ""
    "$binary_path" -v 2>/dev/null || true
    
    if ! check_path "$install_dir"; then
        suggest_add_to_path "$install_dir"
    else
        echo ""
        success "现在可以运行: ${BOLD}ocsession${NC}"
        echo ""
    fi
}

show_help() {
    echo "OpenCode Session Manager 安装脚本 v$INSTALLER_VERSION"
    echo ""
    echo "用法:"
    echo "  curl -sSL https://raw.githubusercontent.com/$REPO/main/install.sh | bash"
    echo ""
    echo "环境变量:"
    echo "  INSTALL_DIR  自定义安装目录（默认: ~/.local/bin）"
    echo "  VERSION      安装指定版本（默认: latest）"
    echo ""
    echo "示例:"
    echo "  # 默认安装"
    echo "  curl -sSL https://raw.githubusercontent.com/$REPO/main/install.sh | bash"
    echo ""
    echo "  # 自定义路径"
    echo "  INSTALL_DIR=/opt/bin curl -sSL ... | bash"
    echo ""
    echo "  # 安装特定版本"
    echo "  VERSION=v1.0.0 curl -sSL ... | bash"
    echo ""
    echo "支持平台:"
    echo "  macOS: arm64 (Apple Silicon), amd64 (Intel)"
    echo "  Linux: amd64 (x86_64), arm64"
}

case "${1:-}" in
    -h|--help|help)
        show_help
        exit 0
        ;;
    -v|--version)
        echo "install.sh version $INSTALLER_VERSION"
        exit 0
        ;;
esac

main