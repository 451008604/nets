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

[//]: # (   <img src="https://img.shields.io/github/issues/451008604/nets.svg" alt="Issues"/>)

[//]: # (   <img src="https://img.shields.io/github/issues-pr/451008604/nets.svg" alt="Pull Requests"/>)

[//]: # (   <img src="https://img.shields.io/github/contributors/451008604/nets.svg" alt="Contributors"/>)

[//]: # (  <img src="https://img.shields.io/github/watchers/451008604/nets.svg?label=Watch" alt="Watchers"/>)

[//]: # (  <img src="https://img.shields.io/github/forks/451008604/nets.svg?label=Fork" alt="Forks"/>)
  <img src="https://img.shields.io/github/stars/451008604/nets.svg?label=Star" alt="Stars"/>
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
| **优雅退出**  | 监听系统信号，并行关闭所有服务与连接                   |

[//]: # (---)

[//]: # (## 📐 架构总览)

[//]: # (<p align="center">)

[//]: # (  <img src="./assets/architecture.svg" alt="NETS Architecture" width="100%"/>)

[//]: # (</p>)

---

## 🔔 环境要求

- **Go ≥ 1.25**

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
)

func main() {
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
进入 `test` 目录，执行 `sh ./start.sh` 启动性能测试

---

## 📄 许可证

[Apache-2.0 License](LICENSE)
