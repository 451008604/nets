package network

import (
	"context"
	"github.com/451008604/nets/config"
	"github.com/451008604/nets/iface"
	"github.com/451008604/nets/logs"
	"io"
	"net"
	"sync"
)

type ConnectionTCP struct {
	Connection
	conn *net.TCPConn // 当前连接对象
}

// 新建连接
func NewConnectionTCP(server iface.IServer, conn *net.TCPConn) *ConnectionTCP {
	c := &ConnectionTCP{}
	c.Server = server
	c.conn = conn
	c.ConnID = server.GetConnMgr().NewConnID()
	c.isClosed = false
	c.MsgHandler = GetInstanceMsgHandler()
	c.exitCtx, c.exitCtxCancel = context.WithCancel(context.Background())
	c.msgBuffChan = make(chan []byte, config.GetGlobalObject().MaxMsgChanLen)
	c.property = make(map[string]interface{})
	c.propertyLock = sync.RWMutex{}
	c.broadcastGroupByID = sync.Map{}
	c.broadcastGroupCh = make(chan iface.IBroadcastData, 1000)
	return c
}

func (c *ConnectionTCP) StartReader() {
	// 获取客户端的消息头信息
	packet := c.Server.DataPacket()
	headData := make([]byte, packet.GetHeadLen())
	if _, err := io.ReadFull(c.conn, headData); err != nil {
		if err != io.EOF {
			logs.PrintLogErr(err)
		}
		c.Stop()
		return
	}
	// 通过消息头获取dataLen和Id
	msgData := packet.Unpack(headData)
	if msgData == nil {
		c.Stop()
		return
	}
	// 通过消息头获取消息body
	if msgData.GetDataLen() > 0 {
		msgData.SetData(make([]byte, msgData.GetDataLen()))
		if _, err := io.ReadFull(c.conn, msgData.GetData()); logs.PrintLogErr(err) {
			c.Stop()
			return
		}
	}

	// 封装请求数据传入处理函数
	req := &Request{conn: c, msg: msgData}
	if config.GetGlobalObject().WorkerPoolSize > 0 {
		c.MsgHandler.SendMsgToTaskQueue(req)
	} else {
		go c.MsgHandler.DoMsgHandler(req)
	}
}

func (c *ConnectionTCP) StartWriter(data []byte) {
	_, err := c.conn.Write(data)
	logs.PrintLogErr(err, string(data))
}

func (c *ConnectionTCP) Start(readerHandler func(), writerHandler func(data []byte)) {
	// 将新建的连接添加到所属Server的连接管理器内
	c.Server.GetConnMgr().Add(c)
	c.JoinBroadcastGroup(c, GetInsBroadcastManager().GetGlobalBroadcast())
	c.Connection.Start(readerHandler, writerHandler)
}

func (c *ConnectionTCP) Stop() {
	if c.isClosed {
		return
	}
	c.Connection.Stop()
	_ = c.conn.Close()
	// 将连接从连接管理器中删除
	c.Server.GetConnMgr().Remove(c)
	c.ExitAllBroadcastGroup()
}

func (c *ConnectionTCP) RemoteAddrStr() string {
	return c.conn.RemoteAddr().String()
}
