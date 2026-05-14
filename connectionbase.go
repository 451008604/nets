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

// 用于生成连接唯一ID
var connIdSeed uint32

func GenerateConnID() string {
	// 1. 获取当前秒级时间戳 (取低 32 位)
	now := uint64(time.Now().Unix())
	// 2. 原子自增获取序列号
	seq := uint64(atomic.AddUint32(&connIdSeed, 1))
	// 3. 组合：时间戳左移 32 位，然后与序列号进行“或”运算
	// [ 32位时间戳 ] [ 32位自增序列 ]
	return strconv.FormatUint((now<<32)|seq, 16)
}

type ConnectionBase struct {
	server        IServer            // 当前Conn所属的Server
	conn          IConnection        // 绑定的连接
	connId        string             // 连接的唯一Id
	msgBuffChan   chan []byte        // 用于任务队列与写协程之间的消息通信
	property      map[string]any     // 连接属性
	propertyMutex sync.RWMutex       // 连接属性读写锁
	isClosed      int32              // 当前连接是否已关闭
	connCtx       context.Context    // 管理连接的上下文
	connCtxCancel context.CancelFunc // 连接关闭信号
	limitingCount int64              // 限流计数
	limitingTimer int64              // 限流计时
	limitingMutex sync.Mutex         // 限流锁
	taskQueue     chan func()        // 等待执行的任务队列
}

func (c *ConnectionBase) Open() {
	defer func() {
		atomic.AddInt32(&c.isClosed, 1)
		GetInstanceConnManager().GetConnClosed(c.conn)

		// 清空属性
		c.propertyMutex.Lock()
		c.property = map[string]any{}
		c.propertyMutex.Unlock()

		// 关闭底层网络连接
		if netConn := c.conn.GetNetConn(); netConn != nil {
			_ = netConn.Close()
		}
		close(c.msgBuffChan)
		c.msgBuffChan = nil
		close(c.taskQueue)
		c.taskQueue = nil

		GetInstanceConnManager().Remove(c.conn)
		GetInstanceServerManager().WaitGroupDone()
	}()

	GetInstanceServerManager().WaitGroupAdd(1)
	GetInstanceConnManager().GetConnOpened(c.conn)

	go c.readHandler()  // 开启读协程
	go c.writeHandler() // 开启写协程

	// 任务协程：处理任务队列，直到上下文取消后排空剩余任务
	for {
		select {
		case <-c.ConnCtxDone():
			return
		case t, ok := <-c.taskQueue:
			if !ok {
				return
			}
			func(taskFun func()) {
				defer GetInstanceMsgHandler().GetErrCapture(c.conn)
				taskFun()
			}(t)
		}
	}
}

func (c *ConnectionBase) readHandler() {
	defer c.Close()
	for {
		select {
		case <-c.ConnCtxDone():
			return
		default:
			// 设置读超时
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

func (c *ConnectionBase) writeHandler() {
	defer c.Close()
	for {
		select {
		case <-c.ConnCtxDone():
			return
		case data, ok := <-c.msgBuffChan:
			if !ok || !c.conn.StartWriter(data) {
				return
			}
		}
	}
}

func (c *ConnectionBase) ConnCtxDone() <-chan struct{} {
	return c.connCtx.Done()
}

func (c *ConnectionBase) Close() {
	// 通知所有协程退出
	c.connCtxCancel()
}

func (c *ConnectionBase) DoTask(task func()) {
	select {
	case <-c.ConnCtxDone():
	case c.taskQueue <- task:
	}
}

func readerTaskHandler(c IConnection, m IMessage) {
	iMsgHandler := GetInstanceMsgHandler()

	// 连接关闭时丢弃后续所有操作
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
	// 限流控制
	if c.FlowControl() {
		fmt.Printf("flowControl RemoteAddress: %v, GetMsgId: %v, GetData: %s\n", c.RemoteAddrStr(), m.GetMsgId(), m.GetData())
		return
	}

	// 过滤器校验
	if iMsgHandler.GetFilter() != nil && !iMsgHandler.GetFilter()(c, m) {
		return
	}

	// 对应的逻辑处理方法
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
	select {
	case <-c.ConnCtxDone():
	case c.msgBuffChan <- msg:
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
