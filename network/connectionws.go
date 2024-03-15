package network

import (
	"context"
	"github.com/451008604/nets/iface"
	"github.com/gorilla/websocket"
	"sync"
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
	c.exitCtx, c.exitCtxCancel = context.WithCancel(context.Background())
	c.msgBuffChan = make(chan []byte, defaultServer.AppConf.MaxMsgChanLen)
	c.property = sync.Map{}
	return c
}

func (c *connectionWS) StartReader() {
	msgType, msgByte, err := c.conn.ReadMessage()
	if err != nil || msgType != websocket.BinaryMessage {
		GetInstanceConnManager().Remove(c)
		return
	}

	msgData := defaultServer.DataPacket.UnPack(msgByte)
	if msgData == nil {
		GetInstanceConnManager().Remove(c)
		return
	}

	for _, data := range msgData {
		// 封装请求数据传入处理函数
		req := &request{conn: c, msg: data}
		if defaultServer.AppConf.WorkerPoolSize > 0 {
			GetInstanceMsgHandler().SendMsgToTaskQueue(req)
		} else {
			go GetInstanceMsgHandler().DoMsgHandler(req)
		}
	}
}

func (c *connectionWS) StartWriter(data []byte) {
	if err := c.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		GetInstanceConnManager().Remove(c)
		return
	}
}

func (c *connectionWS) Start(readerHandler func(), writerHandler func(data []byte)) {
	defer GetInstanceConnManager().Remove(c)

	GetInstanceBroadcastManager().GetGlobalBroadcastGroup().Append(c.GetConnId())
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
