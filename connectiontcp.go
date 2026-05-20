package nets

import (
	"context"
	"net"
)

type connectionTCP struct {
	*ConnectionBase
	conn *net.TCPConn
}

func NewConnectionTCP(server IServer, conn *net.TCPConn) IConnection {
	c := &connectionTCP{
		ConnectionBase: &ConnectionBase{
			server:      server,
			connId:      GenerateConnID(),
			msgBuffChan: make(chan []byte, defaultServer.AppConf.MaxMsgChanLen),
			taskQueue:   make(chan func(), defaultServer.AppConf.WorkerTaskMaxLen),
			property:    map[string]any{},
		},
		conn: conn,
	}
	c.connCtx, c.connCtxCancel = context.WithCancel(context.Background())
	c.ConnectionBase.conn = c
	return c
}

func (c *connectionTCP) GetNetConn() net.Conn {
	return c.conn
}

func (c *connectionTCP) StartReader() bool {
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

	c.DoTask(func() {
		readerTaskHandler(c, msgData)
		PutMessage(msgData)
	})
	return true
}

func (c *connectionTCP) StartWriter(data []byte) bool {
	if _, err := c.conn.Write(data); err != nil {
		return false
	}
	return true
}
