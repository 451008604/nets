package nets

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/protobuf/proto"
)

// 用于生成连接唯一ID
var connIdSeed uint32

type ConnectionBase struct {
	server        IServer            // 当前Conn所属的Server
	conn          IConnection        // 绑定的连接
	connId        string             // 连接的唯一Id
	msgBuffChan   chan []byte        // 用于任务队列与写协程之间的消息通信
	property      map[string]any     // 连接属性
	propertyMutex sync.RWMutex       // 连接属性读写锁
	isClosed      int32              // 当前连接是否已关闭
	exitCtx       context.Context    // 管理连接的上下文
	exitCtxCancel context.CancelFunc // 连接关闭信号
	deadTime      int64              // 读写超时标记
	limitingCount int64              // 限流计数
	limitingTimer int64              // 限流计时
	limitingMutex sync.Mutex         // 限流锁
	taskQueue     chan func()        // 等待执行的任务队列
}

func (c *ConnectionBase) Start() {
	defer func(c *ConnectionBase) {
		c.exitCtxCancel()
		GetInstanceConnManager().ConnOnClosed(c.conn)
		GetInstanceConnManager().Remove(c.conn)
		GetInstanceServerManager().WaitGroupDone()
	}(c)

	GetInstanceServerManager().WaitGroupAdd(1)
	GetInstanceConnManager().ConnOnOpened(c.conn)

	go c.readHandler()  // 开启读协程
	go c.writeHandler() // 开启写协程
	go c.taskHandler()  // 开启任务协程

	tickC := time.Tick(time.Second)
	// 读写超时检测
	for {
		select {
		case <-c.exitCtx.Done():
			return
		case t := <-tickC:
			if t.Unix()-c.deadTime > int64(defaultServer.AppConf.ConnRWTimeOut) {
				return
			}
		}
	}
}

func (c *ConnectionBase) readHandler() {
	defer c.exitCtxCancel()
	for {
		c.deadTime = time.Now().Unix()
		select {
		case <-c.exitCtx.Done():
			return
		default:
			if !c.conn.StartReader() {
				return
			}
		}
	}
}

func (c *ConnectionBase) writeHandler() {
	defer c.exitCtxCancel()
	for {
		c.deadTime = time.Now().Unix()
		select {
		case <-c.exitCtx.Done():
			return
		case data, ok := <-c.msgBuffChan:
			if !ok || !c.conn.StartWriter(data) {
				return
			}
		}
	}
}

func (c *ConnectionBase) taskHandler() {
	defer c.exitCtxCancel()
	for {
		select {
		case <-c.exitCtx.Done():
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

func (c *ConnectionBase) Stop() bool {
	if atomic.AddInt32(&c.isClosed, 1) != 1 {
		return false
	}
	c.exitCtxCancel()
	return true
}

func (c *ConnectionBase) DoTask(task func()) {
	c.taskQueue <- task
}

func readerTaskHandler(c IConnection, m IMessage) {
	iMsgHandler := GetInstanceMsgHandler()

	// 连接关闭时丢弃后续所有操作
	if c.IsClose() {
		return
	}

	router, ok := iMsgHandler.GetApis()[int32(m.GetMsgId())]
	if !ok {
		GetInstanceConnManager().Remove(c)
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
	if c.IsClose() {
		return
	}
	msgByte := c.ProtocolToByte(msgData)
	// 将消息数据封包
	msg := defaultServer.DataPack.Pack(defaultServer.Message(msgId, msgByte))
	if msg == nil {
		return
	}
	// 写入传输通道发送给客户端
	c.msgBuffChan <- msg
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

func (c *ConnectionBase) ProtocolToByte(str proto.Message) []byte {
	var err error
	var marshal []byte

	if defaultServer.AppConf.ProtocolIsJson {
		marshal, err = json.Marshal(str)
	} else {
		marshal, err = proto.Marshal(str)
	}

	if err != nil {
		return []byte{}
	}
	return marshal
}

func (c *ConnectionBase) ByteToProtocol(byte []byte, target proto.Message) error {
	var err error

	if defaultServer.AppConf.ProtocolIsJson {
		err = json.Unmarshal(byte, target)
	} else {
		err = proto.Unmarshal(byte, target)
	}

	if err != nil {
		return err
	}
	return nil
}
