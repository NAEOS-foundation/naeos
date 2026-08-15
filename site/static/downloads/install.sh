#!/bin/sh
# NAEOS installer
# Usage: curl -fsSL https://naeos.dev/install.sh | sh
set -e

REPO="NAEOS-foundation/naeos"
VERSION="${NAEOS_VERSION:-latest}"
BIN_NAME="naeos"
INSTALL_DIR="${NAEOS_INSTALL_DIR:-$HOME/.local/bin}"

arch="$(uname -m)"
os="$(uname -s | tr '[:upper:]' '[:lower:]')"

case "$arch" in
  x86_64|amd64) arch="amd64" ;;
  aarch64|arm64) arch="arm64" ;;
  *) echo "unsupported architecture: $arch" >&2; exit 1 ;;
esac

case "$os" in
  linux|darwin) ;;
  *) echo "unsupported os: $os" >&2; exit 1 ;;
esac

if [ "$VERSION" = "latest" ]; then
  VERSION="$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name"' | head -1 | sed -E 's/.*"([^"]+)".*/\1/')"
fi

FILE_VERSION="${VERSION#v}"
URL="https://github.com/$REPO/releases/download/$VERSION/naeos_${FILE_VERSION}_${os}_${arch}.tar.gz"
echo "Downloading NAEOS $VERSION ($os/$arch)..."

mkdir -p "$INSTALL_DIR"
TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

curl -fsSL "$URL" -o "$TMP/naeos.tar.gz"
tar -xzf "$TMP/naeos.tar.gz" -C "$TMP"
install -m 0755 "$TMP/naeos" "$INSTALL_DIR/$BIN_NAME"

echo "Installed to $INSTALL_DIR/$BIN_NAME"
echo
echo "Add to your PATH if needed: export PATH=\"$INSTALL_DIR:\$PATH\""
if command -v "$BIN_NAME" >/dev/null 2>&1; then
  "$BIN_NAME" version
else
  echo "Run: $INSTALL_DIR/naeos version"
fi
