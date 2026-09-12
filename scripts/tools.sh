#!/bin/sh

set -e

GOLANGCI_LINT_VERSION="2.13.2"
HELMFILE_VERSION="1.7.4"
HELM_DIFF_VERSION="3.15.13"

BIN_DIR="./.bin"

if [ ! -d "$BIN_DIR" ]; then
    echo "Creating $BIN_DIR..."
    mkdir -p "$BIN_DIR"
fi

echo "Downloading golangci-lint v$GOLANGCI_LINT_VERSION..."

curl -L "https://github.com/golangci/golangci-lint/releases/download/v${GOLANGCI_LINT_VERSION}/golangci-lint-${GOLANGCI_LINT_VERSION}-linux-amd64.tar.gz" \
    | tar -xz --strip-components=1 \
        -C "$BIN_DIR" \
        "golangci-lint-${GOLANGCI_LINT_VERSION}-linux-amd64/golangci-lint"

echo "golangci-lint installed to $BIN_DIR/golangci-lint"

echo "Downloading helmfile v$HELMFILE_VERSION..."

curl -L "https://github.com/helmfile/helmfile/releases/download/v${HELMFILE_VERSION}/helmfile_${HELMFILE_VERSION}_linux_amd64.tar.gz" \
    | tar -xz -C "$BIN_DIR" helmfile

echo "helmfile installed to $BIN_DIR/helmfile"

if ! command -v helm > /dev/null 2>&1; then
    echo "helm not found in PATH, skipping helm-diff plugin"
    exit 0
fi

if helm plugin list | grep -q "^diff"; then
    echo "helm-diff plugin already installed"
    exit 0
fi

echo "Installing helm-diff plugin v$HELM_DIFF_VERSION..."

helm plugin install https://github.com/databus23/helm-diff --version "v${HELM_DIFF_VERSION}" --verify=false

echo "helm-diff plugin installed"
