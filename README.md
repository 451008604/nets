<p align="center"><img src="./assets/logo2.webp" alt="" width="200"/></p>

<div align="center">
<img src="https://img.shields.io/github/license/451008604/nets.svg" alt="license"/>
<img src="https://img.shields.io/github/issues/451008604/nets.svg" alt="issues"/>
<img src="https://img.shields.io/github/issues-pr/451008604/nets.svg" alt="issues"/>
<img src="https://img.shields.io/github/contributors/451008604/nets.svg" alt="contributors"/>
</div>
<div align="center">
<img src="https://img.shields.io/github/watchers/451008604/nets.svg?label=Watch" alt="watchers"/>
<img src="https://img.shields.io/github/forks/451008604/nets.svg?label=fork" alt="forks"/>
<img src="https://img.shields.io/github/stars/451008604/nets.svg?label=star" alt="stars"/>
</div>

<!-- TOC -->
* [NETS 简介](#nets-简介)
  * [架构图](#架构图)
* [使用说明](#使用说明)
  * [=> 环境配置](#-环境配置)
  * [=> 快速上手](#-快速上手)
  * [=> Issues](#-issues)
* [致谢](#致谢)
* [许可证](#许可证)
<!-- TOC -->

# NETS 简介

一个追求轻量、性能、实用、可快速上手的网络框架。采用工作池模式，已实现协程复用并且可根据并发数量自动扩容协程池。建立连接只需占用3个协程（1个读协程、1个写协程、1个协程池内的工作协程）

使用面向接口编程和组合设计模式，最大程度提高系统的灵活性、可维护性和可扩展性

**宗旨：拒绝炫技。代码是给人看的，所以首先保证代码要通俗易懂，其次再保证机器运行正常**

现已支持：
* 服务：
  - TCP
  - WebSocket(s)
  - UDP / KCP (🚧进行中)
* 协议：
  - Protocol Buffer
  - JSON
* 功能：
  - [x] 设置连接建立时的前置
  - [x] 设置连接断开时的后置
  - [x] 绑定消息属性
  - [x] 消息处理中间件
  - [x] 自定义编码/解码器
  - [x] 消息业务panic阻断
  - [x] 停服时优雅关闭所有连接
  - [x] 分组广播
  - [x] 全服广播
  - [ ] 广播历史记录 (🚧进行中)

future：  
完善消息广播功能，✅支持创建广播组、✅加入广播组、✅退出广播组、❌广播组解散 (标记不可用，记录保留)

## 架构图
![架构图](./assets/DesignDiagram.drawio.svg)

# 使用说明
### => 环境配置
> Golang >= 1.18

### => 快速上手

- 一个简单的例子
```go
// 启动TCP服务
serverTCP := network.NewServerTCP(nil)
serverTCP.Listen()

// 启动WebSocket服务
serverWS := network.NewServerWS(nil)
serverWS.Listen()

// 阻塞主进程
network.ServerWaitFlag.Wait()
```

- 连接管理器 ( iface.IConnManager ) 的应用  
  **network.GetInstanceConnManager()** 为单例模式，保持全局唯一

```go
connManager := network.GetInstanceConnManager()

// 设置连接建立时的处理
connManager.OnConnOpen(func(conn iface.IConnection) {
    // do something ...
})

// 设置连接断开时的处理
connManager.OnConnClose(func(conn iface.IConnection) {
    // do something ...
})
```

- 消息处理器 ( iface.IMsgHandler ) 的应用

```go
msgHandler := network.GetInstanceMsgHandler()

// 添加一个路由
msgHandler.AddRouter(int32(pb.MSgID_PlayerLogin_Req), func() proto.Message { return &pb.PlayerLoginRequest{} }, func(con iface.IConnection, message proto.Message) {
    // do something ...
})

// 自定义消息过滤器。返回 true 时可正常执行，返回 false 则不会执行路由方法
msgHandler.SetFilter(func(request iface.IRequest, msgData proto.Message) bool {
    // do something ...
    return true
})

// 自定义panic捕获。保障业务逻辑不会导致服务整体崩溃
msgHandler.SetErrCapture(func(request iface.IRequest, r any) {
    // do something ...
})
```

- 广播管理器 ( iface.IBroadcastManager ) 的应用

```go
	broadcastManager := network.GetInstanceBroadcastManager()

```

### => Issues

# 致谢

# 许可证

⚖️[Apache-2.0 license](https://github.com/451008604/nets?tab=Apache-2.0-1-ov-file#)
