# nets 项目架构分析

## 项目概览

- **项目名称**: nets
- **代码规模**: 4212 行（不含测试代码）
- **支持协议**: TCP、WebSocket、HTTP、KCP 四种协议
- **Go 版本**: 1.24.0
- **最后更新**: 2026-03-13

## 核心接口设计

### 基础接口

1. **`IServer`**: 服务器接口
    - `GetServerName() string` - 获取服务器名称
    - `Open()` - 启动服务器

2. **`IConnection`**: 连接接口
    - `Open() / Close() bool` - 启动/停止连接
    - `StartReader() / StartWriter(data []byte)` - 读写协程
    - `DoTask(task func())` - 执行任务
    - `GetConnId() / RemoteAddrStr()` - 连接信息
    - `IsClose() / GetProperty / SetProperty / RemoveProperty()` - 状态与属性管理
    - `SendMsg(msgId, msgData)` - 发送消息（使用 Protobuf）
    - `FlowControl()` - 限流控制
    - `ProtocolToByte() / ByteToProtocol()` - 编解码（JSON/Protobuf）

3. **`IMessage`**: 消息模板接口
    - `GetMsgId() uint16` - 消息ID
    - `GetDataLen() uint16` - 消息长度
    - `GetData() / SetData()` - 消息内容

4. **`IDataPack`**: 封包拆包接口
    - `GetHeadLen() int` - 头部长度（固定4字节）
    - `Pack(msg IMessage) []byte` - 消息封包
    - `UnPack([]byte) IMessage` - 消息拆包

5. **`BaseRouter`**: 路由器模板
    - 结构体字段：
        - `template INewMsgStructTemplate` - 消息结构体工厂
        - `handler IReceiveMsgHandler` - 业务处理函数
    - 方法：`GetNewMsg()` 获取空消息模板，`RunHandler(conn, msg)` 执行处理

## 架构分层

```
┌─────────────────────────────────────────────────────┐
│            ServerManager (服务管理)                  │
│  • 接管所有 IServer 实现                            │
│  • 监听系统信号（SIGINT/SIGTERM/SIGKILL）            │
│  • 优雅停止所有服务                                  │
│  • 统一管理 WaitGroup                                │
└─────────────┬───────────────────────────────────────┘
              │
              ▼
    ┌────────────────────────┐
    │    各协议 Server 实现   │
    ├──────────┬─────────────┤
    │ severtcp │ serverws    │
    │ serverkcp│ serverhttp  │
    └──────────┴─────────────┘
              │
              ▼
┌─────────────────────────────────────────────────────┐
│  ConnectionManager (连接管理)                       │
│  • 基于 shard-map 的分布式连接存储                  │
│  • 连接生命周期管理 (OnOpened / OnClosed)           │
│  • 限流触发 Hook (OnRateLimiting)                   │
└─────────────┬───────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────────────┐
│  ConnectionBase      ← 所有连接的公共实现           │
│    ├── 三大协程：readHandler / writeHandler /      │
│    │          taskHandler                           │
│    ├── 读写超时检测（ConnRWTimeOut 秒）              │
│    ├── 属性管理 (property map[string]any)           │
│    ├── 限流实现 (FlowControl: QPS 阈值)             │
│    └── 任务队列 (chan func())                       │
├─────────────────────────────────────────────────────┤
│    ConnectionTCP  ← TCP 协议（net.TCPConn）         │
│    ConnectionWS   ← WebSocket 协议（gorilla/websocket│
│    ConnectionHTTP ← HTTP 协议（net/http，短连接）   │
│    ConnectionKCP  ← KCP 协协议（xtaci/kcp-go）       │
└─────────────────────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────────────┐
│  MessageHandler (消息处理)                           │
│    ├── 路由表: map[int32]*BaseRouter               │
│    ├── 过滤器: IFilter (conn, msg) -> bool          │
│    └── Panic 捕获: IErrCapture (conn, panicInfo)   │
│         → 调用 defer recover，记录堆栈              │
└───────┬─────────────────────────────────────────────┘
        │
        ▼
┌─────────────────────────────────────────────────────┐
│  readerTaskHandler (消息分发核心)                    │
│  1. 检查连接是否已关闭                               │
│  2. 根据 msgId 查路由表                              │
│  3. 未注册 → 移除连接并返回                          │
│  4. 反序列化: ByteToProtocol(msgData, template)     │
│  5. FlowControl() 限流检查                           │
│  6. 过滤器校验                                       │
│  7. router.RunHandler(conn, msgData)                │
└─────────────────────────────────────────────────────┘
              │
              ▼
┌─────────────────────────────────────────────────────┐
│  DataPack: 封包拆包 [msgId:2B | dataLen:2B | data]   │
│  • 字节序: LittleEndian                             │
│  • 最大包大小限制: MaxPackSize                       │
└─────────────────────────────────────────────────────┘
```

## 核心组件说明

### 1. 配置与自定义入口

**`conf.go`**: `AppConf` 默认配置

- 端口、MaxConn、MaxPackSize
- 是否使用 JSON 协议
- QPS 限流阈值 (MaxFlowSecond = -1 表示禁用)

**`customserver.go`**: 自定义服务器覆盖

- `SetCustomServer(&CustomServer{...})` 覆盖默认配置
- 只覆盖非零值字段（零值不覆盖）
- 合并策略：`mergeStructValues[T]()` 使用反射

### 2. 消息与封包

**`message.go`**: `Message` 结构体

```go
type Message struct {
Id     uint16  // msg_id
DataLen uint16 // data 长度
Data   []byte      // data
}
```

**`datapack.go`**: 默认封包器

- 固定包头：`[msgId:4字节][dataLen:4字节]`（LittleEndian）
- 封包：写入 buff，返回 []byte
- 拆包：读取包头 → 验证长度 → 读取 body
- 最大包检查：超过 `AppConf.MaxPackSize` 返回 nil

### 3. 路由与处理链路

**`router.go`**: `BaseRouter` 路由器

```go
type BaseRouter struct {
template INewMsgStructTemplate
handler  IReceiveMsgHandler
}

// 使用示例
AddRouter(msgId, func () proto.Message { return &TestMsg{} }, handlerFunc)
```

**`msghandler.go`**: `MsgHandler` 消息处理器

- 全局单例（`GetInstanceMsgHandler()`）
- 路由表 `map[int32]*BaseRouter`
- `AddRouter(msgId, template, handler)` - 注册路由
- `SetFilter(IFilter)` - 设置过滤器
- `SetErrCapture(IErrCapture)` - 设置 Panic 捕获
- `GetErrCapture(conn)` - 通过 recover 捕获 executor panic

### 4. 连接模型

**`connectionbase.go`**: 公共连接实现

- **三个并发协程**
    - `readHandler()`: 循环读取数据，调用 `Conn.StartReader()`
    - `writeHandler()`: 从 `msgBuffChan` 读取并发送
    - `taskHandler()`: 从 `taskQueue` 读取并执行任务（带 defer panic 捕获）
- **超时检测**: 每秒检查是否超过 `ConnRWTimeOut` 秒
- **限流**: `FlowControl()` 统计 1 秒内请求数

**具体协议连接**:

- `connectiontcp.go`: TCP 连接，使用 `net.TCPConn`
- `connectionkcp.go`: KCP 连接，使用 `kcp.UDPCONN`
- `connectionws.go`: Websocket，使用 `gorilla/websocket`
- `connectionhttp.go`: HTTP（短连接），直接调用 `readerTaskHandler`

### 5. 连接与服务管理

**`connectionmanager.go`**: 连接管理器

- 使用 `github.com/451008604/shard-map` 分片哈希表
- `Add(conn)` - 添加连接并启动协程
- `Remove(conn)` - 停止连接并删除
- `Get(connId)` - 查找连接
- `RangeConnections(handler)` - 遍历所有连接
- Hook 函数：`OnOpened / OnClosed / OnRateLimiting`

**`servermanager.go`**: 服务管理器

- 管理 `IServer` 接口的实现
- 阻塞等待信号优雅停止
- 创建 WaitGroup 并在连接关闭时 `Done()`
- 停止流程：权限停止所有服务器 → 清空连接 → WaitGroup.Wait()

### 6. 协议差异

| 协议        | 连接类型             | 数据包           | 特点                         |
|-----------|------------------|---------------|----------------------------|
| TCP       | `connectionTCP`  | DataPack      | 长连接，可靠传输                   |
| KCP       | `connectionKCP`  | DataPack      | 基于UDP，抵抗丢包                 |
| WebSocket | `connectionWS`   | BinaryMessage | DataPack 封包                |
| HTTP      | `connectionHTTP` | 无             | 短连接，直接调用 readerTaskHandler |

## 关键特性

### 消息路由

- 按msgId分发到对应的handler
- 未注册msgId直接断开连接

### 数据包处理

- 解决TCP粘包/拆包问题
- 固定4字节包头解析
- 校验最大包大小

### 编码支持

- JSON (默认: false)
- Protobuf (默认: true)
- 通过 `AppConf.ProtocolIsJson` 配置

### 连接管理

- 统一纳入 ConnectionManager
- 基于shard-map的分布式存储
- 生命周期Hook（OnOpened/OnClosed/OnRateLimiting）

### 限流控制

- 基于QPS的限流
- 超过阈值触发 OnRateLimiting 并删除连接
- `MaxFlowSecond = -1` 表示禁用

### Panic捕获

- 通过 defer/recover 捕获handler执行panic
- 记录panic信息和堆栈到IErrCapture
- 不会影响其他连接

## 外部依赖

```
github.com/451008604/nets/internal      # Protobuf 定义
github.com/451008604/shard-map          # 分片哈希表
github.com/gorilla/websocket           # WebSocket
github.com/xtaci/kcp-go                # KCP
google.golang.org/protobuf/proto       # Protobuf
```

### 内部依赖

```
internal/message.pb.go                 # Protobuf 生成的代码
shard-map                              # 分片哈希表实现
```

## 目录结构

```
nets/
├── conf.go                               # 配置定义
├── customserver.go                       # 自定义服务器
├── connectionbase.go                     # 连接公共实现
├── connectionhttp.go                     # HTTP连接
├── connectionkcp.go                     # KCP连接
├── connectiontcp.go                     # TCP连接
├── connectionws.go                      # WebSocket连接
├── connectionmanager.go                 # 连接管理器
├── datapack.go                          # 默认封包器
├── iconnection.go                       # 连接接口
├── idatapack.go                         # 封包接口
├── imessage.go                          # 消息接口
├── iserver.go                           # 服务器接口
├── message.go                           # 消息结构体
├── msghandler.go                        # 消息处理器
├── router.go                            # 路由器模板
├── serverhttp.go                        # HTTP服务器
├── serverkcp.go                        # KCP服务器
├── servermanager.go                     # 服务管理器
├── servertcp.go                        # TCP服务器
└── serverws.go                         # WebSocket服务器
```

## 测试覆盖

所有测试文件已排除分析，包括：

- 单元测试（connection*）。go、msghandler_test.go等）
- 集成测试（测试大规模连接、并发等场景）
- 端到端测试（完整协议栈测试）

## 适用场景

适合用于：

- 后端服务开发
- 实时通讯（即时消息、聊天系统）
- 游戏服务器
- IoT 通信
- 高并发短连接场景（HTTP API）

## 使用示例

```go
// 1. 初始化
GetInstanceMsgHandler().AddRouter(msgId, func () proto.Message { return &Message{} }, handler)
GetInstanceConnManager().SetConnOnOpened(func (conn IConnection) {})
GetInstanceServerManager().StartAll()

// 2. 自定义配置
SetCustomServer(&CustomServer{
AppConf: &AppConf{
ProtocolIsJson: true,
MaxFlowSecond:  1000,
},
})

// 3. 启动服务
server := GetServerTCP()
server.Open()
```

## 代码统计

| 文件类型 | 数量 | 说明                                                               |
|------|----|------------------------------------------------------------------|
| 接口定义 | 5  | iconnection.go, iserver.go, imessage.go, idatapack.go, router.go |
| 实现类  | 14 | 各Server、Connection、Handler等实现                                    |
| 配置类  | 2  | conf.go、customserver.go                                          |
| 工具类  | 1  | datapack.go                                                      |

---

**分析生成时间**: 2026-03-13
**分析工具**: Claude Code
