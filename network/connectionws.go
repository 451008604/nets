package network

import (
	"github.com/451008604/nets/iface"
	"github.com/gorilla/websocket"
	"sync/atomic"
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
	c.property = NewConcurrentMap[any]()
	c.workId = c.connId % defaultServer.AppConf.WorkerPoolSize
	return c
}

func (c *connectionWS) StartReader() bool {
	msgType, msgByte, err := c.conn.ReadMessage()
	if err != nil || msgType != websocket.BinaryMessage {
		atomic.AddUint32(&Flag1, 1)
		GetInstanceConnManager().Remove(c)
		return false
	}

	msgData := defaultServer.DataPacket.UnPack(msgByte)
	if msgData == nil {
		GetInstanceConnManager().Remove(c)
		return false
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
	return true
}

func (c *connectionWS) StartWriter(data []byte) bool {
	if err := c.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		GetInstanceConnManager().Remove(c)
		return false
	}
	return true
}

func (c *connectionWS) Start(readerHandler func() bool, writerHandler func(data []byte) bool) {
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
