<p align="center">
  <img src="./assets/logo2.webp" alt="NETS Logo" width="200"/>
</p>

<div align="center">
  <img src="https://img.shields.io/github/license/451008604/nets.svg" alt="License"/>
  <img src="https://img.shields.io/github/issues/451008604/nets.svg" alt="Issues"/>
  <img src="https://img.shields.io/github/issues-pr/451008604/nets.svg" alt="Pull Requests"/>
  <img src="https://img.shields.io/github/contributors/451008604/nets.svg" alt="Contributors"/>
  <img src="https://img.shields.io/github/watchers/451008604/nets.svg?label=Watch" alt="Watchers"/>
  <img src="https://img.shields.io/github/forks/451008604/nets.svg?label=Fork" alt="Forks"/>
  <img src="https://img.shields.io/github/stars/451008604/nets.svg?label=Star" alt="Stars"/>
</div>

---

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

## 配置速览

> 默认值来源：`conf.go` 中的 `var appConf = &AppConf{...}`。建议在 **第一次调用 `GetServerTCP/WS/HTTP/KCP()` 之前** 先完成配置修改（这些 `GetServer*()` 会把端口/IP 缓存在单例里）。

| 配置项            | 默认值    | 说明                          |
|----------------|--------|-----------------------------|
| TCP 端口         | 17001  | 监听地址：0.0.0.0:17001          |
| WS 端口          | 17002  | 监听地址：0.0.0.0:17002          |
| HTTP 端口        | 17003  | 监听地址：0.0.0.0:17003          |
| KCP 端口         | 17004  | 监听地址：0.0.0.0:17004          |
| ProtocolIsJson | true   | true=JSON 编码，false=Protobuf |
| MaxPackSize    | 4096   | 数据包最大长度（TCP/WS/KCP）         |
| MaxConn        | 100000 | 最大连接数                       |
| MaxFlowSecond  | -1     | 限流阈值（-1 表示关闭）               |
| ConnRWTimeOut  | 30s    | 连接读写超时时间                    |

数据包格式（TCP/WS/KCP）：头 4 字节（`msgId uint16 + dataLen uint16`，小端编码）

> 可通过 `SetCustomServer(&CustomServer{AppConf: ..., DataPack: ..., Message: ...})` 覆盖配置、打包器或消息工厂（`AppConf` 仅"非零值字段"会被合并；例如布尔值 `false` 属于零值，不会覆盖默认值）。

## 快速上手

### 1. 编写业务代码（示例 `main.go`）

```go
package main

import (
    "github.com/451008604/nets"
    "github.com/451008604/nets/internal" // 示例 proto 代码，按需替换
    "google.golang.org/protobuf/proto"
)

func main() {
    // （可选）在第一次调用 GetServer*() 前先修改配置
    // conf := nets.GetServerConf()
    // conf.ProtocolIsJson = true

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
    // 注意：RegisterServer 会阻塞主 goroutine，直到接收到系统信号触发 StopAll。
    nets.GetInstanceServerManager().RegisterServer(
        nets.GetServerTCP(),
        nets.GetServerWS(),
        nets.GetServerHTTP(),
        nets.GetServerKCP(),
    )
}
```

### 2. 运行

```bash
go run main.go
```

默认将开启 TCP/WS/HTTP/KCP 四个端口。

### 3. 测试

```bash
go test ./...   # 运行所有单元/集成测试
```

参考测试文件：`servertcp_test.go`、`serverws_test.go`、`serverkcp_test.go`、`serverhttp_test.go`

## 进阶配置

- **切换 JSON / Proto**：
  ```go
  nets.GetServerConf().ProtocolIsJson = true  // 使用 JSON 编码
  nets.GetServerConf().ProtocolIsJson = false // 使用 Protobuf 编码
  ```

- **限流回调**：
  ```go
  nets.GetInstanceConnManager().SetConnOnRateLimiting(func(conn nets.IConnection) {
      // 触发限流时的处理逻辑
  })
  ```
  当 `MaxFlowSecond != -1` 且触发限流时，会先回调，再立刻断开连接。

- **过滤与错误捕获**：
  ```go
  nets.GetInstanceMsgHandler().SetFilter(func(msgId int32, data []byte) (int32, []byte, bool) {
      // 过滤逻辑，返回 false 表示丢弃消息
  })
  nets.GetInstanceMsgHandler().SetErrCapture(func(err error) {
      // 错误捕获逻辑
  })
  ```

## HTTP（短连接）说明

- HTTP **不使用 DataPack**。`connectionHTTP.SendMsg` 直接把 `ProtocolToByte(msgData)` 写入 response body。
- HTTP 请求体支持两种常用模式：
  1. **msg_id 路由模式**：body 能解析为 `nets.Message`（JSON），且 `msg_id != 0`。此时路由按 `msg_id` 分发，并用 `Message.data` 的 bytes 反序列化为你的业务消息结构体。
  2. **透传/Restful 模式（msg_id=0）**：当 body 无法解析为 `nets.Message`，或解析后 `msg_id == 0`，框架会把原始 body 放入 `Message.Data`，并在连接属性写入 `ConnPropertyHttpReader`/`ConnPropertyHttpWriter`，方便在 handler 中直接访问 `*http.Request` 和 `http.ResponseWriter`。

## Protobuf 示例

示例文件：`internal/message.proto`

生成脚本：

```bash
bash internal/generate_pb.sh
```

## 常见行为与坑点

- **路由未注册会断开连接**：当收到的 `msg_id` 在路由表中不存在时，`readerTaskHandler` 会直接断开连接。
- **触发限流会断开连接**：当 `MaxFlowSecond != -1` 且超过每秒请求数阈值时，会回调限流处理函数，然后移除连接。
- **配置合并规则**：`SetCustomServer` 的 `AppConf` 仅"非零值字段"会被合并。例如布尔值 `false`、整数 `0` 属于零值，不会覆盖默认值。

## 目录结构

```
nets/
├── conf.go              # 配置管理
├── customserver.go      # 自定义服务
├── message.go           # 消息结构
├── datapack.go          # 数据打包器
├── idatapack.go         # 数据打包器接口
├── router.go            # 路由管理
├── msghandler.go        # 消息处理器
├── connectionbase.go    # 连接基类
├── connectiontcp.go     # TCP 连接
├── connectionws.go      # WebSocket 连接
├── connectionhttp.go    # HTTP 连接
├── connectionkcp.go     # KCP 连接
├── connectionmanager.go # 连接管理器
├── servermanager.go     # 服务器管理器
├── servertcp.go         # TCP 服务器
├── serverws.go          # WebSocket 服务器
├── serverhttp.go        # HTTP 服务器
├── serverkcp.go         # KCP 服务器
└── internal/            # 内部示例代码
    ├── message.proto    # Protobuf 定义
    └── generate_pb.sh   # Protobuf 生成脚本
```

## Issues

欢迎提交 Issue 和 Pull Request！

## 致谢

感谢所有贡献者！

## 许可证

[Apache-2.0 license](LICENSE)
