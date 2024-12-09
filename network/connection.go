package network

import (
	"context"
	"encoding/json"
	"github.com/451008604/nets/iface"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"
)

type connection struct {
	server        iface.IServer                           // 当前Conn所属的Server
	connId        int                                     // 当前连接的Id(SessionId)
	msgBuffChan   chan []byte                             // 用于任务队列与写协程之间的消息通信
	property      ConcurrentMap[iface.IConnProperty, any] // 连接属性
	isClosed      bool                                    // 当前连接是否已关闭
	exitCtx       context.Context                         // 管理连接的上下文
	exitCtxCancel context.CancelFunc                      // 连接关闭信号
	limitingCount int64                                   // 限流计数
	limitingTimer int64                                   // 限流计时
	limitingMutex sync.Mutex                              // 限流锁
	taskQueue     chan iface.ITaskTemplate                // 等待执行的任务队列
}

func (c *connection) StartReader() bool { return true }

func (c *connection) StartWriter(_ []byte) bool { return false }

func (c *connection) Start(readerHandler func() bool, writerHandler func(data []byte) bool) {
	defer GetInstanceServerManager().WaitGroupDone()
	defer GetInstanceConnManager().ConnOnClosed(c)

	GetInstanceServerManager().WaitGroupAdd(1)
	GetInstanceConnManager().ConnOnOpened(c)

	// 开启读协程
	go func(c *connection, readerHandler func() bool) {
		for {
			c.exitCtx, c.exitCtxCancel = context.WithTimeout(context.Background(), time.Second*time.Duration(defaultServer.AppConf.ConnRWTimeOut))
			select {
			case <-c.exitCtx.Done():
				return
			default:
				// 调用注册方法处理接收到的消息
				if !readerHandler() {
					GetInstanceConnManager().Remove(c)
					return
				}
			}
		}
	}(c, readerHandler)

	// 开启写协程
	defer close(c.msgBuffChan)
	for {
		c.exitCtx, c.exitCtxCancel = context.WithTimeout(context.Background(), time.Second*time.Duration(defaultServer.AppConf.ConnRWTimeOut))
		select {
		case <-c.exitCtx.Done():
			return
		case data := <-c.msgBuffChan:
			// 调用注册方法写消息给客户端
			if !writerHandler(data) {
				GetInstanceConnManager().Remove(c)
				return
			}
		}
	}
}

func (c *connection) Stop() {
	if c.isClosed {
		return
	}
	c.isClosed = true
	c.exitCtxCancel()
}

func (c *connection) PushTaskQueue(task iface.ITaskTemplate) {
	c.taskQueue <- task
}

func (c *connection) GetConnId() int {
	return c.connId
}

func (c *connection) RemoteAddrStr() string {
	return ""
}

func (c *connection) IsClose() bool {
	return c.isClosed
}

func (c *connection) SendMsg(msgId int32, msgData proto.Message) {
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

func (c *connection) SetProperty(key iface.IConnProperty, value any) {
	c.property.Set(key, value)
}

func (c *connection) GetProperty(key iface.IConnProperty) any {
	if value, ok := c.property.Get(key); ok {
		return value
	} else {
		return nil
	}
}

func (c *connection) RemoveProperty(key iface.IConnProperty) {
	c.property.Remove(key)
}

func (c *connection) FlowControl() bool {
	if defaultServer.AppConf.MaxFlowSecond == 0 {
		return false
	}
	defer c.limitingMutex.Unlock()
	c.limitingMutex.Lock()

	count, interval := int64(defaultServer.AppConf.MaxFlowSecond), int64(1000)
	if c.limitingTimer == 0 {
		c.limitingTimer = time.Now().UnixMilli()
	}
	c.limitingCount++
	if c.limitingCount <= count {
		return false
	}
	now := time.Now().UnixMilli()
	if now-c.limitingTimer < interval {
		GetInstanceConnManager().ConnRateLimiting(c)
		GetInstanceConnManager().Remove(c)
		return true
	}
	c.limitingCount = 1
	c.limitingTimer = now
	return false
}

func (c *connection) ProtocolToByte(str proto.Message) []byte {
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

func (c *connection) ByteToProtocol(byte []byte, target proto.Message) error {
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
