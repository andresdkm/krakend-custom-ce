#!/bin/sh

set -e

echo "=========================================="
echo "KrakenD Plugin Builder"
echo "=========================================="
echo "Go Version: $(go version)"
echo "GOOS: ${GOOS}"
echo "GOARCH: ${GOARCH}"
echo "CGO_ENABLED: ${CGO_ENABLED}"
echo "=========================================="

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Directorio de trabajo
WORKSPACE="/app/src"
OUTPUT_DIR="/app/plugins"

# Verificar si existe el directorio de workspace
if [ ! -d "$WORKSPACE" ]; then
    echo "Error: Workspace directory not found: $WORKSPACE"
    exit 1
fi

cd "$WORKSPACE"
go mod tidy
# Buscar todos los directorios que contengan plugins
find . -maxdepth 2 -name "plugin.go" | while read -r gomod; do
    PLUGIN_DIR=$(dirname "$gomod")
    echo ""
    echo "=========================================="
    echo "Building plugin in: $PLUGIN_DIR"
    echo "=========================================="

    cd "$WORKSPACE/$PLUGIN_DIR"

    # Descargar dependencias
    # echo "Downloading dependencies..."
    # go mod download
    # go mod tidy

    # Compilar el plugin
    PLUGIN_NAME=$(basename "$PLUGIN_DIR")
    echo "Building plugin... $PLUGIN_NAME"

    # Compilar como plugin compartido (.so)
    # Build to a temporary file first to avoid "Resource busy" errors if the file is being read
    # IMPORTANT: -buildmode=plugin is required for Go plugins to be loaded by plugin.Open()
    go build -buildmode=plugin -o "${OUTPUT_DIR}/${PLUGIN_NAME}.so.tmp" .

    if [ $? -eq 0 ]; then
        # Explicitly remove the old file first to handle filesystem locking issues
        if [ -f "${OUTPUT_DIR}/${PLUGIN_NAME}.so" ]; then
            rm -f "${OUTPUT_DIR}/${PLUGIN_NAME}.so"
        fi
        mv "${OUTPUT_DIR}/${PLUGIN_NAME}.so.tmp" "${OUTPUT_DIR}/${PLUGIN_NAME}.so"
        echo "✓ Plugin compiled successfully: ${PLUGIN_NAME}.so"
        ls -lh "${OUTPUT_DIR}/${PLUGIN_NAME}.so"
    else
        echo "✗ Failed to compile plugin: ${PLUGIN_NAME}"
        rm -f "${OUTPUT_DIR}/${PLUGIN_NAME}.so.tmp"
        exit 1
    fi

    cd "$WORKSPACE"
done

echo ""
echo "=========================================="
echo "Build Complete!"
echo "=========================================="
echo "Compiled plugins:"
ls -lh "$OUTPUT_DIR"/*.so 2>/dev/null || echo "No plugins found"
echo "=========================================="