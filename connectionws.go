package nets

import (
	"context"
	"github.com/gorilla/websocket"
	"net"
)

type connectionWS struct {
	*ConnectionBase
	conn *websocket.Conn
}

func NewConnectionWS(server IServer, conn *websocket.Conn) IConnection {
	c := &connectionWS{
		ConnectionBase: &ConnectionBase{
			server:      server,
			connId:      GenerateConnID(),
			msgBuffChan: make(chan []byte, defaultServer.AppConf.MaxMsgChanLen),
			property:    map[string]any{},
		},
		conn: conn,
	}
	c.connCtx, c.connCtxCancel = context.WithCancel(context.Background())
	c.ConnectionBase.conn = c
	return c
}

func (c *connectionWS) GetNetConn() net.Conn {
	return c.conn.NetConn()
}

func (c *connectionWS) StartReader() bool {
	msgType, msgByte, err := c.conn.ReadMessage()
	if err != nil || msgType != websocket.BinaryMessage {
		return false
	}

	for len(msgByte) > 0 {
		msgData := defaultServer.DataPack.UnPack(msgByte)
		if msgData == nil {
			return false
		}
		// Guard against a frame whose declared length exceeds the remaining bytes (avoids slice-out-of-range panic)
		// 防御声明长度超过实际剩余字节的帧（避免切片越界 panic）
		step := int(msgData.GetDataLen()) + defaultServer.DataPack.GetHeadLen()
		if step > len(msgByte) {
			PutMessage(msgData)
			return false
		}
		msgByte = msgByte[step:]

		if !c.DoTask(func() { readerTaskHandler(c, msgData); PutMessage(msgData) }) {
			PutMessage(msgData)
		}
	}
	return true
}

func (c *connectionWS) StartWriter(data []byte) bool {
	if err := c.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return false
	}
	return true
}
