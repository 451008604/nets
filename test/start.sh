#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SERVER_DIR="$SCRIPT_DIR/server"
CLIENT_DIR="$SCRIPT_DIR/client"
SERVER_BIN="$SERVER_DIR/server"
CLIENT_BIN="$CLIENT_DIR/client"

echo "=========================================="
echo "  NETS Automated Test"
echo "=========================================="
echo ""
echo "[Cleanup]"
docker compose down --remove-orphans
echo "Done!"
echo ""
echo "[1/4] Compile..."
GOOS=linux GOARCH=amd64 go build -o "$SERVER_BIN" "$SERVER_DIR/server.go"
GOOS=linux GOARCH=amd64 go build -o "$CLIENT_BIN" "$CLIENT_DIR/client.go"

echo ""
echo "[2/4] Build Docker images..."
cd "$SCRIPT_DIR"
docker compose build

echo ""
echo "[4/4] Run test..."
docker compose up
