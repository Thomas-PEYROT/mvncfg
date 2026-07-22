#!/usr/bin/env bash
set -euo pipefail

# Use the provided shell name, or fall back to the user's $SHELL.
SHELL_NAME="${1:-${SHELL:-}}"
if [ -z "$SHELL_NAME" ]; then
    echo "Cannot detect shell. Please run: ./install.sh <bash|zsh>"
    exit 1
fi
SHELL_NAME=$(basename "$SHELL_NAME")

INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"

VERSION="$(git describe --tags --always --dirty 2>/dev/null || echo dev)"

echo "Building mvncfg ($VERSION)..."
go build -ldflags "-s -w -X main.version=$VERSION" -o "$INSTALL_DIR/mvncfg" ./cmd/mvncfg

echo "Installing $SHELL_NAME completion..."
SHELL="$SHELL_NAME" "$INSTALL_DIR/mvncfg" install-completion

echo ""
echo "mvncfg installed to $INSTALL_DIR/mvncfg"
echo "Reload your shell or run: source ~/.${SHELL_NAME}rc"
