package network

import (
	"context"
	"github.com/451008604/nets/iface"
	"net"
	"time"
)

type connectionTCP struct {
	connection
	conn *net.TCPConn // 当前连接对象
}

func NewConnectionTCP(server iface.IServer, conn *net.TCPConn) iface.IConnection {
	c := &connectionTCP{}
	c.server = server
	c.conn = conn
	c.connId = GetInstanceConnManager().NewConnId()
	c.isClosed = false
	c.msgBuffChan = make(chan []byte, defaultServer.AppConf.MaxMsgChanLen)
	c.property = NewConcurrentStringer[iface.IConnProperty, any]()
	c.taskQueue = GetInstanceWorkerManager().BindTaskQueue(c)
	c.exitCtx, c.exitCtxCancel = context.WithTimeout(context.Background(), time.Second*time.Duration(defaultServer.AppConf.ConnRWTimeOut))
	return c
}

func (c *connectionTCP) StartReader() bool {
	// 获取消息头信息
	msgHead := make([]byte, defaultServer.DataPacket.GetHeadLen())
	if read, err := c.conn.Read(msgHead); err != nil || read < defaultServer.DataPacket.GetHeadLen() {
		return false
	}

	// 解析头信息
	msgData := defaultServer.DataPacket.UnPack(msgHead)
	if msgData == nil {
		return false
	}

	// 解析消息体的内容
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

	// 封装请求数据传入处理函数
	c.PushTaskQueue(msgData)
	return true
}

func (c *connectionTCP) StartWriter(data []byte) bool {
	if _, err := c.conn.Write(data); err != nil {
		return false
	}
	return true
}

func (c *connectionTCP) Start(readerHandler func() bool, writerHandler func(data []byte) bool) {
	defer GetInstanceConnManager().Remove(c)

	c.connection.Start(readerHandler, writerHandler)
}

func (c *connectionTCP) Stop() {
	if c.isClosed {
		return
	}
	c.connection.Stop()
	_ = c.conn.Close()
}

func (c *connectionTCP) RemoteAddrStr() string {
	return c.conn.RemoteAddr().String()
}
