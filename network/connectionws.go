package network

import (
	"github.com/451008604/nets/iface"
	"github.com/gorilla/websocket"
)

type connectionWS struct {
	connection
	conn *websocket.Conn
}

func NewConnectionWS(server iface.IServer, conn *websocket.Conn) iface.IConnection {
	c := &connectionWS{
		connection: connection{
			server:      server,
			connId:      GetInstanceConnManager().NewConnId(),
			msgBuffChan: make(chan []byte, defaultServer.AppConf.MaxMsgChanLen),
			property:    NewConcurrentStringer[iface.IConnProperty, any](),
			taskQueue:   make(chan func(), defaultServer.AppConf.WorkerTaskMaxLen),
		},
		conn: conn,
	}
	return c
}

func (c *connectionWS) StartReader() bool {
	msgType, msgByte, err := c.conn.ReadMessage()
	if err != nil || msgType != websocket.BinaryMessage {
		return false
	}

	for len(msgByte) > 0 {
		msgData := defaultServer.DataPacket.UnPack(msgByte)
		if msgData == nil {
			return false
		}
		msgByte = msgByte[int(msgData.GetDataLen())+defaultServer.DataPacket.GetHeadLen():]

		// 封装请求数据传入处理函数
		c.DoTask(func() { readerTaskHandler(c, msgData) })
	}
	return true
}

func (c *connectionWS) StartWriter(data []byte) bool {
	if err := c.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return false
	}
	return true
}

func (c *connectionWS) Start(readerHandler func() bool, writerHandler func(data []byte) bool) {
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
