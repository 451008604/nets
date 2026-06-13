package nets

import (
	"context"
	"net"
)

type connectionKCP struct {
	*ConnectionBase
	conn net.Conn
}

func NewConnectionKCP(server *serverKCP, conn net.Conn) IConnection {
	msgChanLen := defaultServer.AppConf.MaxMsgChanLen
	if msgChanLen < 0 {
		msgChanLen = 0
	}
	c := &connectionKCP{
		ConnectionBase: &ConnectionBase{
			server:      server,
			connId:      GenerateConnID(),
			msgBuffChan: make(chan []byte, msgChanLen),
			property:    map[string]any{},
		},
		conn: conn,
	}
	c.connCtx, c.connCtxCancel = context.WithCancel(context.Background())
	c.ConnectionBase.conn = c
	return c
}

func (c *connectionKCP) GetNetConn() net.Conn {
	return c.conn
}

func (c *connectionKCP) StartReader() bool {
	msgHead := make([]byte, defaultServer.DataPack.GetHeadLen())
	if read, err := c.conn.Read(msgHead); err != nil || read < defaultServer.DataPack.GetHeadLen() {
		return false
	}

	msgData := defaultServer.DataPack.UnPack(msgHead)
	if msgData == nil {
		return false
	}

	for {
		if len(msgData.GetData()) >= int(msgData.GetDataLen()) {
			break
		}

		bt := make([]byte, int(msgData.GetDataLen())-len(msgData.GetData()))
		read, err := c.conn.Read(bt)
		if err != nil {
			return false
		}
		msgData.SetData(append(msgData.GetData(), bt[:read]...))
	}

	c.submitReaderTask(msgData)
	return true
}

func (c *connectionKCP) StartWriter(data []byte) bool {
	if _, err := c.conn.Write(data); err != nil {
		return false
	}
	return true
}
