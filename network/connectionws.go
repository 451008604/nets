package network

import (
	"context"
	"github.com/451008604/nets/iface"
	"github.com/gorilla/websocket"
)

type connectionWS struct {
	*connectionBase
	conn *websocket.Conn
}

func NewConnectionWS(server iface.IServer, conn *websocket.Conn) iface.IConnection {
	c := &connectionWS{
		connectionBase: &connectionBase{
			server:      server,
			connId:      GetInstanceConnManager().NewConnId(),
			msgBuffChan: make(chan []byte, defaultServer.AppConf.MaxMsgChanLen),
			taskQueue:   make(chan func(), defaultServer.AppConf.WorkerTaskMaxLen),
			property:    NewConcurrentMap[any](),
		},
		conn: conn,
	}
	c.exitCtx, c.exitCtxCancel = context.WithCancel(context.Background())
	c.connectionBase.conn = c
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

func (c *connectionWS) Stop() bool {
	if !c.connectionBase.Stop() {
		return false
	}
	_ = c.conn.Close()
	return true
}

func (c *connectionWS) RemoteAddrStr() string {
	return c.conn.RemoteAddr().String()
}
