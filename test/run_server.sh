#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SERVER_DIR="$SCRIPT_DIR/server"
SERVER_BIN="$SERVER_DIR/server"

echo ""
echo "✅ 编译 server ..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$SERVER_BIN" "$SERVER_DIR/server.go"

echo ""
echo "✅ 构建 server 镜像 ..."
cd "$SCRIPT_DIR"
docker compose up --build server;