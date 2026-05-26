package nets

import (
	"context"
	"encoding/json"
	"fmt"
	"google.golang.org/protobuf/proto"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// Used to Generate Unique Connection ID / 用于生成连接唯一ID
var connIdSeed uint32

func GenerateConnID() string {
	// 1. Get current second-level timestamp (low 32 bits) / 获取当前秒级时间戳 (取低 32 位)
	now := uint64(time.Now().Unix())
	// 2. Atomically increment to get sequence number / 原子自增获取序列号
	seq := uint64(atomic.AddUint32(&connIdSeed, 1))
	// 3. Combine: timestamp shifted left by 32 bits, OR with sequence number / 组合：时间戳左移 32 位，然后与序列号进行“或”运算
	// [ 32-bit Timestamp ] [ 32-bit Auto-increment Sequence / 32位时间戳 ] [ 32位自增序列 ]
	return strconv.FormatUint((now<<32)|seq, 16)
}

type ConnectionBase struct {
	server        IServer            // Current Conn's Server / 当前Conn所属的Server
	conn          IConnection        // Bound Connection / 绑定的连接
	connId        string             // Unique Connection ID / 连接的唯一Id
	msgBuffChan   chan []byte        // Message communication between task queue and write goroutine / 用于任务队列与写协程之间的消息通信
	property      map[string]any     // Connection Properties / 连接属性
	propertyMutex sync.RWMutex       // Connection Property R/W Lock / 连接属性读写锁
	isClosed      int32              // Whether Connection is Closed / 当前连接是否已关闭
	connCtx       context.Context    // Connection Management Context / 管理连接的上下文
	connCtxCancel context.CancelFunc // Connection Close Signal / 连接关闭信号
	limitingCount int64              // Rate Limiting Count / 限流计数
	limitingTimer int64              // Rate Limiting Timer / 限流计时
	limitingMutex sync.Mutex         // Rate Limiting Lock / 限流锁
}

func (c *ConnectionBase) Open() {
	defer func() {
		c.Close()
		GetInstanceConnManager().GetConnClosed(c.conn)

		// Clear Properties / 清空属性
		c.propertyMutex.Lock()
		c.property = map[string]any{}
		c.propertyMutex.Unlock()

		// Close Underlying Network Connection / 关闭底层网络连接
		if netConn := c.conn.GetNetConn(); netConn != nil {
			_ = netConn.Close()
		}
		close(c.msgBuffChan)
		c.msgBuffChan = nil

		GetInstanceConnManager().Remove(c.conn)
		GetInstanceServerManager().WaitGroupDone()
	}()

	GetInstanceServerManager().WaitGroupAdd(1)
	GetInstanceConnManager().GetConnOpened(c.conn)

	// Start Read Goroutine / 开启读协程
	go c.readHandler()

	// Start Write Goroutine / 开启写协程
	for {
		select {
		case <-c.ConnCtx().Done():
			return
		case data, ok := <-c.msgBuffChan:
			if !ok || !c.conn.StartWriter(data) {
				return
			}
		}
	}
}

func (c *ConnectionBase) readHandler() {
	defer c.Close()
	for {
		select {
		case <-c.ConnCtx().Done():
			return
		default:
			// Set Read Timeout / 设置读超时
			deadline := time.Now().Add(time.Duration(defaultServer.AppConf.ConnRWTimeOut) * time.Second)
			if netConn := c.conn.GetNetConn(); netConn != nil && netConn.SetReadDeadline(deadline) != nil {
				return
			}
			if !c.conn.StartReader() {
				return
			}
		}
	}
}

func (c *ConnectionBase) ConnCtx() context.Context {
	return c.connCtx
}

func (c *ConnectionBase) Close() {
	atomic.AddInt32(&c.isClosed, 1)
	// Notify all goroutines to exit / 通知所有协程退出
	c.connCtxCancel()
}

func (c *ConnectionBase) DoTask(task func()) {
	if c.IsClose() {
		return
	}
	// Hash connId to workerId, ensure all handlers of same connection execute on same worker
	// 将 connId 哈希为 workerId，确保同一连接的所有 handler 在同一 worker 上执行
	pool := GetInstanceWorkerPool()
	err := pool.SubmitWithWorkerCtx(c.ConnCtx(), func() {
		defer GetInstanceMsgHandler().GetErrCapture(c.conn)
		task()
	}, pool.HashWorkerId(c.connId))
	if err != nil {
		fmt.Printf("pool do task err:%v\n", err)
		return
	}
}

func readerTaskHandler(c IConnection, m IMessage) {
	iMsgHandler := GetInstanceMsgHandler()

	// Discard all subsequent operations when connection closes / 连接关闭时丢弃后续所有操作
	if c.IsClose() {
		return
	}

	router, ok := iMsgHandler.GetApis()[int32(m.GetMsgId())]
	if !ok {
		return
	}

	msgData := router.GetNewMsg()
	if m.GetMsgId() != 0 {
		if err := c.ByteToProtocol(m.GetData(), msgData); err != nil {
			fmt.Printf("api msgId %v parsing %s error %v\n", m.GetMsgId(), m.GetData(), err)
			return
		}
	} else {
		msgData = m.(*Message)
	}
	// Rate Limiting Control / 限流控制
	if c.FlowControl() {
		fmt.Printf("flowControl RemoteAddress: %v, GetMsgId: %v, GetData: %s\n", c.RemoteAddrStr(), m.GetMsgId(), m.GetData())
		return
	}

	// Filter Validation / 过滤器校验
	if iMsgHandler.GetFilter() != nil && !iMsgHandler.GetFilter()(c, m) {
		return
	}

	// Corresponding Logic Handler / 对应的逻辑处理方法
	router.RunHandler(c, msgData)
}

func (c *ConnectionBase) GetConnId() string {
	return c.connId
}

func (c *ConnectionBase) IsClose() bool {
	return atomic.LoadInt32(&c.isClosed) != 0
}

func (c *ConnectionBase) GetProperty(key string) any {
	c.propertyMutex.RLock()
	defer c.propertyMutex.RUnlock()
	return c.property[key]
}

func (c *ConnectionBase) SetProperty(key string, value any) {
	c.propertyMutex.Lock()
	defer c.propertyMutex.Unlock()
	c.property[key] = value
}

func (c *ConnectionBase) RemoveProperty(key string) {
	c.propertyMutex.Lock()
	defer c.propertyMutex.Unlock()
	delete(c.property, key)
}

func (c *ConnectionBase) SendMsg(msgId int32, msgData proto.Message) {
	msgByte := c.ProtocolToByte(msgData)
	packMsg := GetMessage()
	packMsg.Id = uint16(msgId)
	packMsg.Data = msgByte
	msg := defaultServer.DataPack.Pack(packMsg)
	PutMessage(packMsg)
	if msg == nil {
		return
	}
	// Non-blocking check connection closed first / 先非阻塞检查连接是否关闭
	select {
	case <-c.ConnCtx().Done():
		return
	default:
	}
	// Then try to send message / 再尝试发送消息
	select {
	case c.msgBuffChan <- msg:
	case <-c.ConnCtx().Done():
	}
}

func (c *ConnectionBase) FlowControl() (b bool) {
	defer c.limitingMutex.Unlock()
	c.limitingMutex.Lock()

	defer func() {
		if b {
			GetInstanceConnManager().ConnRateLimiting(c.conn)
			GetInstanceConnManager().Remove(c.conn)
		}
	}()

	count := int64(defaultServer.AppConf.MaxFlowSecond)
	if count == -1 {
		return false
	}

	if c.limitingTimer == 0 {
		c.limitingTimer = time.Now().UnixMilli()
	}
	c.limitingCount++
	if c.limitingCount <= count {
		return false
	}
	now := time.Now().UnixMilli()
	if now-c.limitingTimer < int64(1000) {
		return true
	}
	c.limitingCount = 1
	c.limitingTimer = now
	return false
}

func (c *ConnectionBase) RemoteAddrStr() string {
	return c.conn.GetNetConn().RemoteAddr().String()
}

var (
	protoMarshalOnce sync.Once
	protoMarshal     func(proto.Message) ([]byte, error)
	protoUnmarshal   func([]byte, proto.Message) error
)

func initProtoFuncs() {
	protoMarshalOnce.Do(func() {
		if defaultServer.AppConf.ProtocolIsJson {
			protoMarshal = func(m proto.Message) ([]byte, error) { return json.Marshal(m) }
			protoUnmarshal = func(b []byte, m proto.Message) error { return json.Unmarshal(b, m) }
		} else {
			protoMarshal = proto.MarshalOptions{}.Marshal
			protoUnmarshal = proto.UnmarshalOptions{}.Unmarshal
		}
	})
}

func (c *ConnectionBase) ProtocolToByte(str proto.Message) []byte {
	initProtoFuncs()
	marshal, err := protoMarshal(str)
	if err != nil {
		return []byte{}
	}
	return marshal
}

func (c *ConnectionBase) ByteToProtocol(byte []byte, target proto.Message) error {
	initProtoFuncs()
	return protoUnmarshal(byte, target)
}
