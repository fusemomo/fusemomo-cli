#!/bin/bash
set -e

# Fusemomo CLI - Installer
# Usage: curl -sL https://.../install.sh | bash

OWNER="fusemomo"
REPO="fusemomo-cli"
BINARY="fusemomo"

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "${OS}" in
  linux*)   OS='linux';;
  darwin*)  OS='darwin';;
  msys*|cygwin*|mingw*) OS='windows';;
  *)        echo "Error: Unsupported OS ${OS}"; exit 1;;
esac

# Detect Arch
ARCH="$(uname -m)"
case "${ARCH}" in
  x86_64) ARCH='amd64';;
  arm64|aarch64) ARCH='arm64';;
  *)      echo "Error: Unsupported architecture ${ARCH}"; exit 1;;
esac

# Get latest version from GitHub
VERSION=$(curl -s "https://api.github.com/repos/${OWNER}/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "${VERSION}" ]; then
  echo "Error: Could not determine latest version."
  exit 1
fi

echo "Downloading ${BINARY} ${VERSION} for ${OS}/${ARCH}..."

# Construct download URL (matching .goreleaser.yaml archive template)
# Template: fusemomo_{{.Version}}_{{.Os}}_{{.Arch}}
FORMAT="tar.gz"
if [ "${OS}" = "windows" ]; then
  FORMAT="zip"
fi

URL="https://github.com/${OWNER}/${REPO}/releases/download/${VERSION}/${BINARY}_${VERSION}_${OS}_${ARCH}.${FORMAT}"

TMP_DIR=$(mktemp -d)
curl -sL "${URL}" -o "${TMP_DIR}/archive.${FORMAT}"

# Extract
cd "${TMP_DIR}"
if [ "${FORMAT}" = "zip" ]; then
  unzip -q "archive.zip"
else
  tar -xzf "archive.tar.gz"
fi

# Install
INSTALL_DIR="/usr/local/bin"
if [ ! -w "${INSTALL_DIR}" ]; then
  echo "Installation requires sudo permissions."
  sudo mv "${BINARY}" "${INSTALL_DIR}/${BINARY}"
else
  mv "${BINARY}" "${INSTALL_DIR}/${BINARY}"
fi

echo "Successfully installed ${BINARY} to ${INSTALL_DIR}/${BINARY}"
${BINARY} version
