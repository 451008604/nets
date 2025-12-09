package network

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/451008604/nets/iface"
	"google.golang.org/protobuf/proto"
)

type connectionBase struct {
	server        iface.IServer              // 当前Conn所属的Server
	conn          iface.IConnection          // 绑定的连接
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

func (c *connectionBase) Start(readerHandler func() bool, writerHandler func(data []byte) bool) {
	defer GetInstanceServerManager().WaitGroupDone()
	defer GetInstanceConnManager().Remove(c.conn)
	defer GetInstanceConnManager().ConnOnClosed(c.conn)
	defer c.exitCtxCancel()

	GetInstanceServerManager().WaitGroupAdd(1)
	GetInstanceConnManager().ConnOnOpened(c.conn)

	// 开启读协程
	go func(c *connectionBase, readerHandler func() bool) {
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
	go func(c *connectionBase, writerHandler func(data []byte) bool) {
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
	go func(c *connectionBase) {
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

func (c *connectionBase) Stop() bool {
	if c.isClosed {
		return false
	}
	c.isClosed = true
	c.exitCtxCancel()
	return true
}

func (c *connectionBase) DoTask(task func()) {
	c.taskQueue <- task
}

func readerTaskHandler(c iface.IConnection, m iface.IMessage) {
	iMsgHandler := GetInstanceMsgHandler()
	defer iMsgHandler.GetErrCapture(c, m)

	// 连接关闭时丢弃后续所有操作
	if c.IsClose() {
		return
	}

	router, ok := iMsgHandler.GetApis()[int32(m.GetMsgId())]
	if !ok {
		return
	}

	msgData := router.GetNewMsg()
	if err := c.ByteToProtocol(m.GetData(), msgData); err != nil {
		fmt.Printf("api msgId %v parsing %v error %v\n", m.GetMsgId(), m.GetData(), err)
		return
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

func (c *connectionBase) GetConnId() int {
	return c.connId
}

func (c *connectionBase) RemoteAddrStr() string {
	return ""
}

func (c *connectionBase) IsClose() bool {
	return c.isClosed
}

func (c *connectionBase) SendMsg(msgId int32, msgData proto.Message) {
	if c.isClosed {
		return
	}
	msgByte := c.ProtocolToByte(msgData)
	// 将消息数据封包
	msg := defaultServer.DataPacket.Pack(NewMsgPackage(msgId, msgByte))
	if msg == nil {
		return
	}
	// 写入传输通道发送给客户端
	c.msgBuffChan <- msg
}

func (c *connectionBase) FlowControl() (b bool) {
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

func (c *connectionBase) ProtocolToByte(str proto.Message) []byte {
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

func (c *connectionBase) ByteToProtocol(byte []byte, target proto.Message) error {
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
func ConnPropertySet(c *connectionBase, key string, value any) {
	c.property.Set(key, value)
}

// 获取连接属性
func ConnPropertyGet[T any](c *connectionBase, key string) T {
	var t T
	if value, ok := c.property.Get(key); ok {
		if v, ok2 := value.(T); ok2 {
			return v
		}
	}
	return t
}

// 删除连接属性
func ConnPropertyRemove(c *connectionBase, key string) {
	c.property.Remove(key)
}
