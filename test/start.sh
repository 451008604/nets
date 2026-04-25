#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
SERVER_DIR="$SCRIPT_DIR/server"
CLIENT_DIR="$SCRIPT_DIR/client"
SERVER_BIN="$SERVER_DIR/server"
CLIENT_BIN="$CLIENT_DIR/client"
TEST_DURATION=30

echo "=========================================="
echo "  NETS Automated Test"
echo "=========================================="
echo ""
echo "[1/4] Compile..."
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o "$SERVER_BIN" "$SERVER_DIR/server.go"
CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o "$CLIENT_BIN" "$CLIENT_DIR/client.go"

echo ""
echo "[2/4] Build Docker images..."
cd "$SCRIPT_DIR"
docker compose build

echo ""
echo "[3/4] Start server..."
docker compose up -d server

echo "  Wait for healthy..."
for i in {1..20}; do
    sleep 1
    STATUS=$(docker inspect test-server-1 --format='{{.State.Health.Status}}' 2>/dev/null || echo "starting")
    [ "$STATUS" = "healthy" ] && break
    echo "  Status: $STATUS (try $i/20)"
done

echo ""
echo "[4/4] Run test..."
docker compose up -d client

sleep "$TEST_DURATION"

echo ""
echo "=========================================="
echo "  RESULTS"
echo "=========================================="
docker compose logs server
docker compose logs client

echo ""
echo "[Cleanup]"
docker compose down --remove-orphans
echo "Done!"