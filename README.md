# NETS

一个面向服务端的多协议网络框架，统一抽象 TCP / WebSocket / HTTP / KCP，提供路由、连接与生命周期管理，支持 JSON / Protobuf 编解码和限流，适合作为后端网关或游戏/实时服务的底层网络库。

## 特性概览
- **多协议统一**：同一套接口启动 TCP、WS、HTTP、KCP 服务，复用消息路由与连接管理。
- **路由解耦**：`MsgHandler` 通过 `AddRouter(msgId, tmpl, handler)` 绑定消息结构体工厂与业务处理函数。
- **编解码可选**：`AppConf.ProtocolIsJson` 在 JSON 与 Protobuf 间切换；可自定义 `DataPack` 和消息工厂。
- **连接与流控**：`ConnectionBase` 拆分读/写/任务协程，支持属性存取、超时、基于 QPS 的限流与回调。
- **生命周期管理**：`ServerManager` 并行启动多服务，监听系统信号后优雅关闭全部连接。

## 环境要求
- Go ≥ 1.24（`go.mod`：`go 1.24.0`，`toolchain go1.24.11`）

## 安装与构建
```bash
git clone https://github.com/451008604/nets.git
cd nets
go mod tidy
```

## 配置速览（`conf.go` 默认值）
- 端口：TCP `17001`，WS `17002`，HTTP `17003`，KCP `17004`
- JSON/Proto：`ProtocolIsJson`（默认 `false` 使用 Protobuf）
- 数据包：头 4 字节（msgId uint16 + dataLen uint16，小端），`MaxPackSize` 默认 4096
- 连接/限流：`MaxConn` 10000，`MaxFlowSecond` 1000，`ConnRWTimeOut` 5s

> 可通过 `SetCustomServer(&CustomServer{AppConf: ..., DataPack: ..., Message: ...})` 覆盖配置、打包器或消息工厂（仅非零/非空字段会被合并）。

## 快速上手
下面示例展示如何注册路由并同时启动四种协议服务。

```go
package main

import (
    "github.com/451008604/nets"
    "github.com/451008604/nets/internal" // 示例 proto 代码，按需替换
    "google.golang.org/protobuf/proto"
)

func main() {
    // 注册业务路由
    nets.GetInstanceMsgHandler().AddRouter(
        int32(internal.Test_MsgId_Test_Echo),
        func() proto.Message { return &internal.Test_EchoRequest{} },
        func(conn nets.IConnection, m proto.Message) {
            req := m.(*internal.Test_EchoRequest)
            conn.SendMsg(int32(internal.Test_MsgId_Test_Echo), &internal.Test_EchoResponse{Message: req.Message})
        },
    )

    // 启动多协议服务（默认监听 0.0.0.0）
    nets.GetInstanceServerManager().RegisterServer(
        nets.GetServerTCP(),
        nets.GetServerWS(),
        nets.GetServerHTTP(),
        nets.GetServerKCP(),
    )
}
```

运行：
```bash
go run main.go
```
默认将开启 TCP/WS/HTTP/KCP 四个端口。

## 进阶配置
- **切换 JSON / Proto**：`GetServerConf().ProtocolIsJson = true`
- **限流回调**：实现 `ConnRateLimiting`，在 `ConnectionManager` Hook 中接收通知
- **过滤与错误捕获**：`MsgHandler.SetFilter(filter)`，`MsgHandler.SetErrCapture(capture)`

## 测试与示例
- 运行全部测试：`go test ./...`
- 参考客户端示例：`client_ws_test.go`、`client_kcp_test.go`
- Protobuf 示例：`internal/message.proto`，生成脚本 `internal/generate_pb.sh`

## 目录提示（关键 Go 文件）
- 配置与自定义：`conf.go`，`customserver.go`
- 消息与打包：`message.go`，`datapack.go`，`idatapack.go`
- 路由与处理：`router.go`，`msghandler.go`
- 连接基类与协议：`connectionbase.go`，`connectiontcp.go`，`connectionws.go`，`connectionhttp.go`，`connectionkcp.go`
- 服务器管理：`servermanager.go`，`servertcp.go`，`serverws.go`，`serverhttp.go`，`serverkcp.go`
