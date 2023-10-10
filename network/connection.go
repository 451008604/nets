package network

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/451008604/socketServerFrame/config"
	"github.com/451008604/socketServerFrame/iface"
	"github.com/451008604/socketServerFrame/logs"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
	"sync"
)

type Connection struct {
	Server          iface.IServer          // 当前Conn所属的Server
	ConnID          int                    // 当前连接的ID（SessionID）
	isClosed        bool                   // 当前连接是否已关闭
	MsgHandler      iface.IMsgHandler      // 消息管理MsgId和对应处理函数的消息管理模块
	exitCtx         context.Context        // 管理连接的上下文
	exitCtxCancel   context.CancelFunc     // 连接关闭信号
	msgBuffChan     chan []byte            // 用于读、写两个goroutine之间的消息通信
	property        map[string]interface{} // 连接属性
	propertyLock    sync.RWMutex           // 连接属性读写锁
	player          interface{}            // 玩家数据
	notifyGroupByID sync.Map               // 通知组ID列表
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
		case <-c.exitCtx.Done():
			return
		default:
			// 处理广播消息
			c.notifyGroupByID.Range(func(key, value any) bool {
				notify := value.(iface.INotify)
				if v := notify.GetNotifyCtx().Value("notify"); v != nil {
					data := v.(*NotifyData)
					c.SendMsg(data.MsgID, data.MsgData)
				}
				return true
			})
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

func (c *Connection) RemoteAddrStr() string {
	return ""
}

func (c *Connection) SendMsg(msgId pb.MSgID, msgData proto.Message) {
	msgByte := c.ProtocolToByte(msgData)
	if c.isClosed {
		logs.PrintLogInfo(fmt.Sprintf("连接已关闭导致消息发送失败 -> msgId:%v\tdata:%v", msgId, msgByte))
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

func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = value
}

func (c *Connection) GetProperty(key string) interface{} {
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

func (c *Connection) SetPlayer(player interface{}) {
	c.player = player
}

func (c *Connection) GetPlayer() interface{} {
	return c.player
}

func (c *Connection) JoinNotifyGroup(conn iface.IConnection, group iface.INotify) {
	c.notifyGroupByID.Store(group.GetGroupID(), group)
	group.SetNotifyTarget(conn)
}

func (c *Connection) ExitNotifyGroupByID(groupID int64) {
	if value, loaded := c.notifyGroupByID.LoadAndDelete(groupID); loaded {
		group := value.(iface.INotify)
		group.DelNotifyTarget(c.GetConnID())
	}
}

func (c *Connection) ExitAllNotifyGroup() {
	c.notifyGroupByID.Range(func(key, value any) bool {
		group := value.(iface.INotify)
		group.DelNotifyTarget(c.GetConnID())
		return true
	})
	c.notifyGroupByID = sync.Map{}
}

func (c *Connection) ProtocolToByte(str proto.Message) []byte {
	var err error
	var marshal []byte

	if config.GetGlobalObject().ProtocolIsJson {
		marshal, err = json.Marshal(str)
	} else {
		marshal, err = proto.Marshal(str)
	}

	if err != nil {
		logs.PrintLogErr(err)
		return []byte{}
	}
	return marshal
}

func (c *Connection) ByteToProtocol(byte []byte, target proto.Message) error {
	var err error

	if config.GetGlobalObject().ProtocolIsJson {
		err = json.Unmarshal(byte, target)
	} else {
		err = proto.Unmarshal(byte, target)
	}

	if err != nil {
		logs.PrintLogErr(err)
		return err
	}
	return nil
}
