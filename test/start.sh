#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SERVER_DIR="$SCRIPT_DIR/server"
CLIENT_DIR="$SCRIPT_DIR/client"
SERVER_BIN="$SERVER_DIR/server"
CLIENT_BIN="$CLIENT_DIR/client"

# 清理函数
cleanup() {
    echo ""
    echo "✅ 收到退出信号，清理镜像和容器..."
    docker compose down --remove-orphans --rmi all
    echo "✅ 清理完毕"
    exit 0
}
# 捕获 Ctrl+C 和终止信号
trap cleanup SIGINT SIGTERM

echo "=========================================="
echo "  NETS Automated Test"
echo "=========================================="

echo ""
echo "✅ 初始化测试环境..."
docker compose down --remove-orphans --rmi all

echo ""
echo "✅ 编译server和client..."
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$SERVER_BIN" "$SERVER_DIR/server.go"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "$CLIENT_BIN" "$CLIENT_DIR/client.go"

echo ""
echo "✅ 构建Docker镜像并启动..."
cd "$SCRIPT_DIR"
docker compose up --build -d

# 循环测试
ROUND=1
while true; do
    echo "✅ 开始第 $ROUND 轮测试..."
    docker compose up client --no-deps
    ((ROUND++))
done
