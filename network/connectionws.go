package network

import (
	"context"
	"errors"
	"github.com/451008604/nets/config"
	"github.com/451008604/nets/iface"
	"github.com/451008604/nets/logs"
	"github.com/gorilla/websocket"
	"io"
	"sync"
)

type ConnectionWS struct {
	Connection
	conn *websocket.Conn
}

func NewConnectionWS(server iface.IServer, conn *websocket.Conn) *ConnectionWS {
	c := &ConnectionWS{}
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

func (c *ConnectionWS) StartReader() {
	msgType, msgByte, err := c.conn.ReadMessage()
	if err != nil || msgType != websocket.BinaryMessage {
		if !errors.As(err, &io.ErrUnexpectedEOF) {
			logs.PrintLogErr(err)
		}
		c.Stop()
		return
	}

	packet := c.Server.DataPacket()
	msgData := packet.Unpack(msgByte)
	if msgData == nil {
		c.Stop()
		return
	}
	if msgData.GetDataLen() > 0 {
		msgData.SetData(msgByte[packet.GetHeadLen() : packet.GetHeadLen()+int(msgData.GetDataLen())])
	}

	// 封装请求数据传入处理函数
	req := &Request{conn: c, msg: msgData}
	if config.GetGlobalObject().WorkerPoolSize > 0 {
		c.MsgHandler.SendMsgToTaskQueue(req)
	} else {
		go c.MsgHandler.DoMsgHandler(req)
	}
}

func (c *ConnectionWS) StartWriter(data []byte) {
	err := c.conn.WriteMessage(websocket.BinaryMessage, data)
	logs.PrintLogErr(err, string(data))
}

func (c *ConnectionWS) Start(readerHandler func(), writerHandler func(data []byte)) {
	// 将新建的连接添加到所属Server的连接管理器内
	c.Server.GetConnMgr().Add(c)
	c.JoinBroadcastGroup(c, GetInsBroadcastManager().GetGlobalBroadcast())
	c.Connection.Start(readerHandler, writerHandler)
}

func (c *ConnectionWS) Stop() {
	if c.isClosed {
		return
	}
	c.Connection.Stop()
	_ = c.conn.Close()
	// 将连接从连接管理器中删除
	c.Server.GetConnMgr().Remove(c)
	c.ExitAllBroadcastGroup()
}

func (c *ConnectionWS) RemoteAddrStr() string {
	return c.conn.RemoteAddr().String()
}
