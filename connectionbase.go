package nets

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"google.golang.org/protobuf/proto"
)

type ConnectionBase struct {
	server        IServer                    // 当前Conn所属的Server
	conn          IConnection                // 绑定的连接
	connId        int                        // 当前连接的Id(SessionId)
	msgBuffChan   chan []byte                // 用于任务队列与写协程之间的消息通信
	property      ConcurrentMap[string, any] // 连接属性
	isClosed      bool                       // 当前连接是否已关闭
	exitCtx       context.Context            // 管理连接的上下文
	exitCtxCancel context.CancelFunc         // 连接关闭信号
	deadTime      int64                      // 读写超时标记
	limitingCount int64                      // 限流计数
	limitingTimer int64                      // 限流计时
	limitingMutex sync.Mutex                 // 限流锁
	taskQueue     chan func()                // 等待执行的任务队列
}

func (c *ConnectionBase) Start(readerHandler func() bool, writerHandler func(data []byte) bool) {
	defer GetInstanceServerManager().WaitGroupDone()
	defer GetInstanceConnManager().Remove(c.conn)
	defer GetInstanceConnManager().ConnOnClosed(c.conn)
	defer c.exitCtxCancel()

	GetInstanceServerManager().WaitGroupAdd(1)
	GetInstanceConnManager().ConnOnOpened(c.conn)

	// 开启读协程
	go func(c *ConnectionBase, readerHandler func() bool) {
		defer c.exitCtxCancel()
		for {
			c.deadTime = time.Now().Unix()
			select {
			case <-c.exitCtx.Done():
				return
			default:
				// 调用注册方法处理接收到的消息
				if !readerHandler() {
					return
				}
			}
		}
	}(c, readerHandler)

	// 开启写协程
	go func(c *ConnectionBase, writerHandler func(data []byte) bool) {
		defer c.exitCtxCancel()
		for {
			c.deadTime = time.Now().Unix()
			select {
			case <-c.exitCtx.Done():
				return
			case data, ok := <-c.msgBuffChan:
				// 调用注册方法写消息给客户端
				if !ok || !writerHandler(data) {
					return
				}
			}
		}
	}(c, writerHandler)

	// 开启任务协程
	go func(c *ConnectionBase) {
		defer c.exitCtxCancel()
		for {
			select {
			case <-c.exitCtx.Done():
				return
			case taskFun, ok := <-c.taskQueue:
				if !ok {
					return
				}
				taskFun()
			}
		}
	}(c)

	// 读写超时检测
	for {
		select {
		case <-c.exitCtx.Done():
			return
		case t := <-time.Tick(time.Second):
			if t.Unix()-c.deadTime > int64(defaultServer.AppConf.ConnRWTimeOut) {
				return
			}
		}
	}
}

func (c *ConnectionBase) Stop() bool {
	if c.isClosed {
		return false
	}
	c.isClosed = true
	c.exitCtxCancel()
	return true
}

func (c *ConnectionBase) DoTask(task func()) {
	c.taskQueue <- task
}

func readerTaskHandler(c IConnection, m IMessage) {
	iMsgHandler := GetInstanceMsgHandler()
	defer iMsgHandler.GetErrCapture(c, m)

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

func (c *ConnectionBase) GetConnId() int {
	return c.connId
}

func (c *ConnectionBase) IsClose() bool {
	return c.isClosed
}

func (c *ConnectionBase) GetProperty() any {
	return c.property
}

func (c *ConnectionBase) SendMsg(msgId int32, msgData proto.Message) {
	if c.isClosed {
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
	if count == 0 {
		return true
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

// 设置连接属性
func ConnPropertySet(c IConnection, key string, value any) {
	c.GetProperty().(ConcurrentMap[string, any]).Set(key, value)
}

// 获取连接属性
func ConnPropertyGet[T any](c IConnection, key string) T {
	var t T
	if value, ok := c.GetProperty().(ConcurrentMap[string, any]).Get(key); ok {
		if v, ok2 := value.(T); ok2 {
			return v
		}
	}
	return t
}

// 删除连接属性
func ConnPropertyRemove(c IConnection, key string) {
	c.GetProperty().(ConcurrentMap[string, any]).Remove(key)
}
