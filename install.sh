#!/usr/bin/env sh
# Conjure installer — Linux & macOS
# Usage: curl -sfL https://raw.githubusercontent.com/WizardOpsTech/conjure/main/install.sh | sh
set -e

# Detect OS
OS=$(uname -s)
case "$OS" in
  Linux)  OS_SLUG="Linux" ;;
  Darwin) OS_SLUG="Darwin" ;;
  *)
    echo "Unsupported OS: $OS"
    echo "Windows users: run this in PowerShell instead:"
    echo "  irm https://raw.githubusercontent.com/WizardOpsTech/conjure/main/install.ps1 | iex"
    exit 1
    ;;
esac

# Detect architecture
ARCH=$(uname -m)
case "$ARCH" in
  x86_64)        ARCH_SLUG="x86_64" ;;
  aarch64|arm64) ARCH_SLUG="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH"
    exit 1
    ;;
esac

# Fetch latest release tag from the GitHub API
echo "Fetching latest Conjure release..."
TAG=$(curl -sfL "https://api.github.com/repos/WizardOpsTech/conjure/releases/latest" \
  | grep '"tag_name"' \
  | head -1 \
  | cut -d'"' -f4)

if [ -z "$TAG" ]; then
  echo "Failed to fetch release info. Check your internet connection and try again."
  exit 1
fi

ARCHIVE="conjure_${TAG}_${OS_SLUG}_${ARCH_SLUG}.tar.gz"
URL="https://github.com/WizardOpsTech/conjure/releases/download/${TAG}/${ARCHIVE}"

echo "Installing conjure ${TAG} for ${OS_SLUG}/${ARCH_SLUG}..."

# Download and extract the binary to a temp directory
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

curl -sfL "$URL" | tar xzf - -C "$TMP" conjure
chmod +x "$TMP/conjure"

# Install to /usr/local/bin — use sudo if the directory is not writable
INSTALL_DIR="/usr/local/bin"
if [ -w "$INSTALL_DIR" ]; then
  mv "$TMP/conjure" "$INSTALL_DIR/conjure"
else
  echo "Installing to $INSTALL_DIR requires elevated permissions..."
  sudo mv "$TMP/conjure" "$INSTALL_DIR/conjure"
fi

echo ""
echo "conjure ${TAG} installed to ${INSTALL_DIR}/conjure"
echo "Run: conjure --version"
