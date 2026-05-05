#!/bin/sh
set -e

REPO="rohankewal/pom"
BIN="pom"
INSTALL_DIR="/usr/local/bin"

# Detect OS
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  darwin) OS="darwin" ;;
  linux)  OS="linux"  ;;
  *)
    echo "Unsupported OS: $OS"
    exit 1
    ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64 | amd64) ARCH="amd64" ;;
  aarch64 | arm64) ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

# Get latest version
VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
  | grep '"tag_name"' | head -1 | cut -d'"' -f4)

if [ -z "$VERSION" ]; then
  echo "Could not determine latest version."
  exit 1
fi

ARCHIVE="${BIN}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"

echo "Installing pom ${VERSION} (${OS}/${ARCH})..."

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

curl -fsSL "$URL" -o "$TMP/$ARCHIVE"
tar -xzf "$TMP/$ARCHIVE" -C "$TMP"

# Install — try /usr/local/bin, fall back to ~/bin
if [ -w "$INSTALL_DIR" ]; then
  install -m755 "$TMP/$BIN" "$INSTALL_DIR/$BIN"
else
  mkdir -p "$HOME/bin"
  install -m755 "$TMP/$BIN" "$HOME/bin/$BIN"
  INSTALL_DIR="$HOME/bin"
  echo "Note: installed to ~/bin — make sure it's in your PATH."
fi

echo "pom ${VERSION} installed to ${INSTALL_DIR}/${BIN}"
echo "Run 'pom --version' to verify."
