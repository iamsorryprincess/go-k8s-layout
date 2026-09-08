#!/bin/sh

set -e

VERSION="2.13.2"
BIN_DIR="./.bin"

if [ ! -d "$BIN_DIR" ]; then
    echo "Creating $BIN_DIR..."
    mkdir -p "$BIN_DIR"
fi

echo "Downloading golangci-lint v$VERSION..."

curl -L "https://github.com/golangci/golangci-lint/releases/download/v${VERSION}/golangci-lint-${VERSION}-linux-amd64.tar.gz" \
    | tar -xz --strip-components=1 \
        -C "$BIN_DIR" \
        "golangci-lint-${VERSION}-linux-amd64/golangci-lint"

echo "golangci-lint installed to $BIN_DIR/golangci-lint"
