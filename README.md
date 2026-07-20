[//]: # (<p align="center">)

[//]: # (    <img src="" alt="NETS Logo" width="200"/>)

[//]: # (</p>)

<h1 align="center">NETS</h1>

<p align="center">
  <strong>多协议 all-in-one 网络框架</strong><br/>
  一站式启动 TCP / WebSocket / HTTP / KCP 服务，专注路由、连接与生命周期管理<br />
  稳定、快速、安全
</p>

<div align="center">
  <img src="https://img.shields.io/github/license/451008604/nets.svg" alt="License"/>
  <img src="https://img.shields.io/github/stars/451008604/nets.svg?label=Star" alt="Stars"/>

[//]: # (   <img src="https://img.shields.io/github/issues/451008604/nets.svg" alt="Issues"/>)

[//]: # (   <img src="https://img.shields.io/github/issues-pr/451008604/nets.svg" alt="Pull Requests"/>)

[//]: # (   <img src="https://img.shields.io/github/contributors/451008604/nets.svg" alt="Contributors"/>)

[//]: # (  <img src="https://img.shields.io/github/watchers/451008604/nets.svg?label=Watch" alt="Watchers"/>)

[//]: # (  <img src="https://img.shields.io/github/forks/451008604/nets.svg?label=Fork" alt="Forks"/>)
</div>

---

## ✨ 特性

| 特性        | 说明                                   |
|:----------|:-------------------------------------|
| **多协议统一** | TCP、WebSocket、HTTP、KCP 四种协议共用路由与连接管理 |
| **路由解耦**  | 通过 `msgId` 绑定消息工厂与业务处理函数，扩展性强        |
| **编解码可选** | JSON / Protobuf 一键切换，支持自定义 DataPack  |
| **连接管理**  | 分片哈希表存储，读写任务三协程分离，属性存取               |
| **限流控制**  | 基于 QPS 的连接限流，支持回调与自动断开               |
| **优雅退出**  | 监听系统信号，在进程退出前并行关闭所有服务与连接             |
| **高性能**   | 4核8G单台机器 CPS(每秒新建连接数) 可达 10,000+            |

---

## 📐 架构总览

```mermaid
flowchart TB
    Client["客户端<br/>请求数据包"]
    Protocol["多协议服务<br/>TCP / WebSocket / HTTP / KCP"]

    subgraph RequestPaths["请求处理路径"]
        direction LR

        subgraph LongPath["TCP / WebSocket / KCP"]
            direction TB
            LongConn["Connection<br/>长连接读包"]
            Unpack["IDataPack.UnPack<br/>拆包"]
            Workers["WorkerPool<br/>按连接分发任务"]

            LongConn --> Unpack
            Unpack --> Workers
        end

        subgraph HTTPPath["HTTP"]
            direction TB
            HTTPConn["HTTP Connection<br/>读取请求体"]
        end
    end

    MsgHandler["MsgHandler<br/>路由与过滤"]
    Host["宿主应用<br/>业务 Handler"]

    Client -->|"请求数据包"| Protocol
    Protocol -->|"TCP / WebSocket / KCP"| LongConn
    Protocol -->|"HTTP"| HTTPConn

    HTTPConn -->|"短请求直接分发"| MsgHandler
    Workers --> MsgHandler
    MsgHandler --> Host

    subgraph Managers["管理组件"]
        direction LR
        ServerManager["ServerManager<br/>服务生命周期"]
        ConnManager["ConnectionManager<br/>连接管理与 Hook"]
    end

    Host -.->|"注册服务"| ServerManager
    Host -.->|"注册路由和 Hook"| MsgHandler
    ServerManager -.->|"管理"| Protocol
    ServerManager -.->|"退出清理"| ConnManager
    ServerManager -.->|"退出停止"| Workers
    ConnManager -.->|"管理长连接"| LongConn
```

- **实线**表示请求数据包的实际处理流向，**虚线**表示注册、管理和生命周期控制关系。
- TCP、WebSocket、KCP 请求经过 Connection、`IDataPack.UnPack` 和 WorkerPool，再进入统一消息处理链。
- HTTP 请求读取并解析 Body 后直接进入统一消息处理链，不经过长连接 WorkerPool。
- `MsgHandler` 完成路由查找、协议反序列化、限流和 Filter 校验，最后调用宿主注册的业务 Handler。

---

## 🔔 环境要求

- **Go ≥ 1.21.1**

---

## 🚀 快速上手

1. 安装

```bash
go get github.com/451008604/nets
```

2. 启动

```go
package main

import (
    "github.com/451008604/nets"
    "github.com/451008604/nets/internal"
    "google.golang.org/protobuf/proto"
	"log/slog"
	"os"
)

func main() {
	// 可选：注入宿主项目的 slog.Logger，SDK 会保留 Handler、级别和 With 属性。
	conf := nets.GetServerConf()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "game-gateway")
	nets.SetCustomServer(&nets.CustomServer{AppConf: conf, Logger: logger})

    // 1. 注册路由
    nets.GetInstanceMsgHandler().AddRouter(int32(internal.Test_MsgId_Test_Echo), func() proto.Message { return &internal.Test_EchoRequest{} }, func(conn nets.IConnection, message proto.Message) {
        // 获取请求数据
        msgReq, _ := message.(*internal.Test_EchoRequest)
        // 构造响应数据
        msgRes := &internal.Test_EchoResponse{}
        // 业务处理完毕发送响应数据
        defer conn.SendMsg(int32(internal.Test_MsgId_Test_Echo), msgRes)

        // ...处理逻辑，并设置响应数据...
        msgRes.Message = msgReq.GetMessage()
    })

    // 2. 启动服务(阻塞主协程) 
    nets.GetInstanceServerManager().RegisterServer(nets.GetServerHTTP(), nets.GetServerKCP(), nets.GetServerTCP(), nets.GetServerWS())
}
```

3. [详细用法参考](./test/server/server.go)

---

## 🔧 分布式压测

性能测试工具位于 `test` 目录，采用 `docker compose` 编排 `1个server + N个client` 模拟海量客户端并发测试。具体测试配置项位于 [docker-compose.yml](./test/docker-compose.yml)  
进入 `test` 目录
- 执行 `sh run_all.sh` 启动完整性能测试
- 执行 `sh run_client.sh` 单独启动客户端

```shell
# 查看内存占用
go tool pprof -http=:8080 http://localhost:16060/debug/pprof/heap

# 查看 server 日志
docker compose logs server -f
```

---

## 📄 许可证

[Apache-2.0 License](LICENSE)
