#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SERVER_DIR="$SCRIPT_DIR/server"
CLIENT_DIR="$SCRIPT_DIR/client"
SERVER_BIN="$SERVER_DIR/server"
CLIENT_BIN="$CLIENT_DIR/client"

#echo ""
#echo "✅ 初始化测试环境..."
#docker compose down --remove-orphans --rmi all

echo ""
echo "✅ 编译server和client..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$SERVER_BIN" "$SERVER_DIR/server.go"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$CLIENT_BIN" "$CLIENT_DIR/client.go"

echo ""
echo "✅ 构建 Server 镜像并启动..."
cd "$SCRIPT_DIR"
docker compose up --build server
