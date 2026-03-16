#!/bin/bash
set -e

OWNER="fusemomo"
REPO="fusemomo-cli"
ARCHIVE="fusemomo-cli"
BINARY="fusemomo-cli"
INSTALL_NAME="fusemomo"

# Detect OS
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
case "${OS}" in
  linux*)               OS='linux';;
  darwin*)              OS='darwin';;
  msys*|cygwin*|mingw*) OS='windows';;
  *)        echo "Error: Unsupported OS ${OS}"; exit 1;;
esac

# Detect Arch
ARCH="$(uname -m)"
case "${ARCH}" in
  x86_64)        ARCH='amd64';;
  arm64|aarch64) ARCH='arm64';;
  *)      echo "Error: Unsupported architecture ${ARCH}"; exit 1;;
esac

# Get latest version tag
VERSION=$(curl -s "https://api.github.com/repos/${OWNER}/${REPO}/releases/latest" \
  | grep '"tag_name":' \
  | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "${VERSION}" ]; then
  echo "Error: Could not determine latest version."
  exit 1
fi

VERSION_STRIPPED="${VERSION#v}"

URL="https://github.com/${OWNER}/${REPO}/releases/download/${VERSION}/${ARCHIVE}_${VERSION_STRIPPED}_${OS}_${ARCH}.tar.gz"

echo "Downloading ${INSTALL_NAME} ${VERSION} for ${OS}/${ARCH}..."

TMP_DIR=$(mktemp -d)
curl -sL "${URL}" -o "${TMP_DIR}/archive.tar.gz"

cd "${TMP_DIR}"
tar -xzf "archive.tar.gz"

INSTALL_DIR="/usr/local/bin"
if [ ! -w "${INSTALL_DIR}" ]; then
  sudo mv "${BINARY}" "${INSTALL_DIR}/${INSTALL_NAME}"
else
  mv "${BINARY}" "${INSTALL_DIR}/${INSTALL_NAME}"
fi

echo "Successfully installed ${INSTALL_NAME} ${VERSION} to ${INSTALL_DIR}/${INSTALL_NAME}"
${INSTALL_NAME} version