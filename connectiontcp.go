package nets

import (
	"net"
)

type connectionTCP struct {
	*ConnectionBase
	conn *net.TCPConn
}

func NewConnectionTCP(server IServer, conn *net.TCPConn) IConnection {
	c := &connectionTCP{
		ConnectionBase: newConnectionBase(server, managerForServer(server).ctx, true),
		conn:           conn,
	}
	c.ConnectionBase.conn = c
	return c
}

func (c *connectionTCP) GetNetConn() net.Conn {
	return c.conn
}

func (c *connectionTCP) StartReader() bool { return readStreamPacket(c, c.conn) }

func (c *connectionTCP) StartWriter(data []byte) bool { return writeStream(c.conn, data) }
