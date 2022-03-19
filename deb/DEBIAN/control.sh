#!/bin/bash
PACKAGE_NAME="Hidra"
VERSION=$(git describe --tags --abbrev=0)
VERSION=${VERSION/v/}

ARCH="$1"

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

echo "Package: $PACKAGE_NAME" > "$SCRIPT_DIR/control"
echo "Version: $VERSION" >> "$SCRIPT_DIR/control"
echo "Architecture: $ARCH" >> "$SCRIPT_DIR/control"
echo "Maintainer: Hidra Team <hola@josecarlos.me>" >> "$SCRIPT_DIR/control"
echo "Description: Hidra allows you to monitor your services by declaring scenarios in a simple and sequential way" >> "$SCRIPT_DIR/control"

