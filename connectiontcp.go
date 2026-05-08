package nets

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type connectionTCP struct {
	*ConnectionBase
	conn   *net.TCPConn
	writer *bufio.Writer
}

func NewConnectionTCP(server IServer, conn *net.TCPConn) IConnection {
	c := &connectionTCP{
		ConnectionBase: &ConnectionBase{
			server:        server,
			connId:        fmt.Sprintf("%X-%.10v", time.Now().Unix(), atomic.AddUint32(&connIdSeed, 1)),
			msgBuffChan:   make(chan []byte, defaultServer.AppConf.MaxMsgChanLen),
			taskQueue:     make(chan func(), defaultServer.AppConf.WorkerTaskMaxLen),
			property:      map[string]any{},
			propertyMutex: sync.RWMutex{},
		},
		conn: conn,
	}
	c.writer = bufio.NewWriterSize(conn, 64*1024)
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
	if _, err := c.writer.Write(data); err != nil {
		return false
	}
	if err := c.writer.Flush(); err != nil {
		return false
	}
	return true
}

func (c *connectionTCP) RemoteAddrStr() string {
	return c.conn.RemoteAddr().String()
}
