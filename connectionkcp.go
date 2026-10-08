package nets

import (
	"net"
)

type connectionKCP struct {
	*ConnectionBase
	conn net.Conn
}

func NewConnectionKCP(server IServer, conn net.Conn) IConnection {
	c := &connectionKCP{
		ConnectionBase: newConnectionBase(server, managerForServer(server).ctx, true),
		conn:           conn,
	}
	c.ConnectionBase.conn = c
	return c
}

func (c *connectionKCP) GetNetConn() net.Conn {
	return c.conn
}

func (c *connectionKCP) StartReader() bool { return readStreamPacket(c, c.conn) }

func (c *connectionKCP) StartWriter(data []byte) bool { return writeStream(c.conn, data) }
