package nets

import (
	"github.com/gorilla/websocket"
	"net"
)

type connectionWS struct {
	*ConnectionBase
	conn *websocket.Conn
}

func NewConnectionWS(server IServer, conn *websocket.Conn) IConnection {
	c := &connectionWS{
		ConnectionBase: newConnectionBase(server, managerForServer(server).ctx, true),
		conn:           conn,
	}
	c.ConnectionBase.conn = c
	conn.SetReadLimit(int64(c.manager.config.AppConf.MaxPackSize) + int64(c.manager.config.DataPack.GetHeadLen()))
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
		msgData := c.manager.config.DataPack.UnPack(msgByte)
		if msgData == nil {
			return false
		}
		// Guard against a frame whose declared length exceeds the remaining bytes (avoids slice-out-of-range panic)
		// 防御声明长度超过实际剩余字节的帧（避免切片越界 panic）
		step := int(msgData.GetDataLen()) + c.manager.config.DataPack.GetHeadLen()
		if step <= 0 || step > len(msgByte) || uint(msgData.GetDataLen()) > c.manager.config.AppConf.MaxPackSize {
			PutMessage(msgData)
			return false
		}
		msgByte = msgByte[step:]

		if !submitReadTask(c, func() { defer PutMessage(msgData); readerTaskHandler(c, msgData) }) {
			PutMessage(msgData)
			return false
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
