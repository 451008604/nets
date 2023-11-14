package network

import (
	"context"
	"encoding/json"
	"github.com/451008604/nets/config"
	"github.com/451008604/nets/iface"
	"google.golang.org/protobuf/proto"
	"sync"
)

type Connection struct {
	Server             iface.IServer             // 当前Conn所属的Server
	ConnID             int                       // 当前连接的ID（SessionID）
	isClosed           bool                      // 当前连接是否已关闭
	exitCtx            context.Context           // 管理连接的上下文
	exitCtxCancel      context.CancelFunc        // 连接关闭信号
	msgBuffChan        chan []byte               // 用于读、写两个goroutine之间的消息通信
	property           map[string]any            // 连接属性
	propertyLock       sync.RWMutex              // 连接属性读写锁
	broadcastGroupByID sync.Map                  // 广播组列表
	broadcastGroupCh   chan iface.IBroadcastData // 广播数据通道
}

func (c *Connection) StartReader() {}

func (c *Connection) StartWriter(_ []byte) {}

func (c *Connection) Start(readerHandler func(), writerHandler func(data []byte)) {
	defer c.Stop()

	// 开启读协程
	go func(c *Connection, readerHandler func()) {
		for {
			select {
			default:
				// 调用注册方法处理接收到的消息
				readerHandler()
			case <-c.exitCtx.Done():
				return
			}
		}
	}(c, readerHandler)

	// 开启写协程
	for {
		select {
		case data := <-c.msgBuffChan:
			// 调用注册方法写消息给客户端
			writerHandler(data)

		case data := <-c.broadcastGroupCh:
			c.SendMsg(data.MsgID(), data.MsgData())

		case <-c.exitCtx.Done():
			return
		}
	}
}

func (c *Connection) Stop() {
	if c.isClosed {
		return
	}
	c.isClosed = true
	// 通知关闭该连接的监听
	c.exitCtxCancel()

	// 关闭该连接管道
	close(c.msgBuffChan)
}

func (c *Connection) GetConnID() int {
	return c.ConnID
}

func (c *Connection) SetNotifyGroupCh(broadcastGroupCh iface.IBroadcastData) {
	c.broadcastGroupCh <- broadcastGroupCh
}

func (c *Connection) RemoteAddrStr() string {
	return ""
}

func (c *Connection) SendMsg(msgId int32, msgData proto.Message) {
	msgByte := c.ProtocolToByte(msgData)
	if c.isClosed {
		return
	}

	// 将消息数据封包
	msg := c.Server.DataPacket().Pack(NewMsgPackage(msgId, msgByte))
	if msg == nil {
		return
	}
	// 写入传输通道发送给客户端
	c.msgBuffChan <- msg
}

func (c *Connection) SetProperty(key string, value any) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = value
}

func (c *Connection) GetProperty(key string) any {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if value, ok := c.property[key]; ok {
		return value
	} else {
		return nil
	}
}

func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}

func (c *Connection) JoinBroadcastGroup(conn iface.IConnection, group iface.IBroadcast) {
	c.broadcastGroupByID.Store(group.GetGroupID(), group)
	group.SetBroadcastTarget(conn)
}

func (c *Connection) ExitBroadcastGroupByID(groupID int64) {
	if value, loaded := c.broadcastGroupByID.LoadAndDelete(groupID); loaded {
		value.(iface.IBroadcast).DelBroadcastTarget(c.GetConnID())
	}
}

func (c *Connection) ExitAllBroadcastGroup() {
	c.broadcastGroupByID.Range(func(key, value any) bool {
		value.(iface.IBroadcast).DelBroadcastTarget(c.GetConnID())
		return true
	})
	c.broadcastGroupByID = sync.Map{}
}

func (c *Connection) ProtocolToByte(str proto.Message) []byte {
	var err error
	var marshal []byte

	if config.GetServerConf().ProtocolIsJson {
		marshal, err = json.Marshal(str)
	} else {
		marshal, err = proto.Marshal(str)
	}

	if err != nil {
		return []byte{}
	}
	return marshal
}

func (c *Connection) ByteToProtocol(byte []byte, target proto.Message) error {
	var err error

	if config.GetServerConf().ProtocolIsJson {
		err = json.Unmarshal(byte, target)
	} else {
		err = proto.Unmarshal(byte, target)
	}

	if err != nil {
		return err
	}
	return nil
}
