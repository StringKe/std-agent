#!/usr/bin/env sh
# std-agent 安装脚本 (macOS / Linux)
# 用法:
#   curl -fsSL https://raw.githubusercontent.com/StringKe/std-agent/main/install.sh | sh
# 可选环境变量:
#   STD_AGENT_OWNER       GitHub owner (默认 StringKe)
#   STD_AGENT_REPO        仓库名 (默认 std-agent)
#   STD_AGENT_VERSION     版本 tag (如 v0.1.0；默认 latest)
#   STD_AGENT_INSTALL_DIR 安装目录 (默认 $HOME/.local/bin)

set -eu

OWNER="${STD_AGENT_OWNER:-StringKe}"
REPO="${STD_AGENT_REPO:-std-agent}"
VERSION="${STD_AGENT_VERSION:-latest}"
INSTALL_DIR="${STD_AGENT_INSTALL_DIR:-${HOME}/.local/bin}"
BIN_NAME="stdagent"

log()  { printf '\033[1;34m[install]\033[0m %s\n' "$*"; }
warn() { printf '\033[1;33m[install]\033[0m %s\n' "$*" >&2; }
die()  { printf '\033[1;31m[install]\033[0m %s\n' "$*" >&2; exit 1; }

require() {
  command -v "$1" >/dev/null 2>&1 || die "未找到必需命令: $1"
}

require curl
require tar
require uname
require mktemp
if command -v sha256sum >/dev/null 2>&1; then
  SHA_CMD="sha256sum"
elif command -v shasum >/dev/null 2>&1; then
  SHA_CMD="shasum -a 256"
else
  die "未找到 sha256sum 或 shasum"
fi

os="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "$os" in
  linux)  os=linux ;;
  darwin) os=darwin ;;
  *) die "不支持的操作系统: $os" ;;
esac

arch="$(uname -m)"
case "$arch" in
  x86_64|amd64) arch=amd64 ;;
  arm64|aarch64) arch=arm64 ;;
  *) die "不支持的架构: $arch" ;;
esac

if [ "$VERSION" = "latest" ]; then
  log "解析最新版本"
  api_url="https://api.github.com/repos/${OWNER}/${REPO}/releases/latest"
  VERSION="$(curl -fsSL "$api_url" \
    | sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' \
    | head -n1)"
  [ -n "$VERSION" ] || die "无法解析最新版本，请显式指定 STD_AGENT_VERSION"
fi

ver_no_v="${VERSION#v}"
archive="${REPO}_${ver_no_v}_${os}_${arch}.tar.gz"
download_base="https://github.com/${OWNER}/${REPO}/releases/download/${VERSION}"
archive_url="${download_base}/${archive}"
checksum_url="${download_base}/checksums.txt"

tmp="$(mktemp -d 2>/dev/null || mktemp -d -t std-agent-install)"
trap 'rm -rf "$tmp"' EXIT INT TERM

log "下载 $archive_url"
curl -fsSL --retry 3 -o "$tmp/$archive" "$archive_url"

log "下载并校验 checksum"
curl -fsSL --retry 3 -o "$tmp/checksums.txt" "$checksum_url"
expected="$(grep " ${archive}\$" "$tmp/checksums.txt" | awk '{print $1}')"
[ -n "$expected" ] || die "checksums.txt 中找不到 $archive"
actual="$(cd "$tmp" && $SHA_CMD "$archive" | awk '{print $1}')"
[ "$expected" = "$actual" ] || die "checksum 不匹配 expected=$expected actual=$actual"

log "解包"
tar -xzf "$tmp/$archive" -C "$tmp"
[ -f "$tmp/$BIN_NAME" ] || die "归档中未找到二进制 $BIN_NAME"

mkdir -p "$INSTALL_DIR"
mv "$tmp/$BIN_NAME" "$INSTALL_DIR/$BIN_NAME"
chmod +x "$INSTALL_DIR/$BIN_NAME"

log "已安装 $INSTALL_DIR/$BIN_NAME ($VERSION)"
case ":$PATH:" in
  *":$INSTALL_DIR:"*) ;;
  *) warn "目录未在 PATH 中，请将下行加入 ~/.profile 或对应 rc 文件: export PATH=\"$INSTALL_DIR:\$PATH\"" ;;
esac
