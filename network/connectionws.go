package network

import (
	"context"
	"github.com/451008604/nets/iface"
	"github.com/gorilla/websocket"
)

type connectionWS struct {
	connection
	conn *websocket.Conn
}

func NewConnectionWS(server iface.IServer, conn *websocket.Conn) iface.IConnection {
	c := &connectionWS{}
	c.server = server
	c.conn = conn
	c.connId = GetInstanceConnManager().NewConnId()
	c.isClosed = false
	c.msgBuffChan = make(chan []byte, defaultServer.AppConf.MaxMsgChanLen)
	c.property = NewConcurrentStringer[iface.IConnProperty, any]()
	c.taskQueue = GetInstanceWorkerManager().BindTaskQueue(c)
	c.exitCtx, c.exitCtxCancel = context.WithCancel(context.Background())
	return c
}

func (c *connectionWS) StartReader() bool {
	msgType, msgByte, err := c.conn.ReadMessage()
	if err != nil || msgType != websocket.BinaryMessage {
		return false
	}

	msgData := defaultServer.DataPacket.UnPack(msgByte)
	if msgData == nil {
		return false
	}

	// 封装请求数据传入处理函数
	c.PushTaskQueue(msgData)
	return true
}

func (c *connectionWS) StartWriter(data []byte) bool {
	if err := c.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return false
	}
	return true
}

func (c *connectionWS) Start(readerHandler func() bool, writerHandler func(data []byte) bool) {
	defer GetInstanceConnManager().Remove(c)

	c.connection.Start(readerHandler, writerHandler)
}

func (c *connectionWS) Stop() {
	if c.isClosed {
		return
	}
	c.connection.Stop()
	_ = c.conn.Close()
}

func (c *connectionWS) RemoteAddrStr() string {
	return c.conn.RemoteAddr().String()
}
