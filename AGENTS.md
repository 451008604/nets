# AGENTS.md

## Build & Test

- `go mod tidy` - Download/update dependencies
- `go test ./...` - Run all tests
- `bash internal/generate_pb.sh` - Generate protobuf code

## Configuration Timing

- **MUST configure BEFORE first `GetServer*()` call** - ports/IP cached in singletons

## Config Merge Rules

- `SetCustomServer(&CustomServer{AppConf: ...})` only merges non-zero fields
- Zero values (false, 0) do NOT override defaults

## Important Behaviors

- Unknown msgId → connection disconnected
- Rate limiting (MaxFlowSecond != -1) → callback + disconnect

## Key Files

- `conf.go` - Default config (ports, MaxConn, ProtocolIsJson, MaxFlowSecond)
- `customserver.go` - Custom server setup
- `internal/message.proto` - Protobuf definitions