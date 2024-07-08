package network

import (
	"github.com/451008604/nets/iface"
	"net"
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
	c.property = NewConcurrentMap[any]()
	c.workId = c.connId
	if defaultServer.AppConf.WorkerPoolSize != 0 {
		c.workId %= defaultServer.AppConf.WorkerPoolSize
	}
	return c
}

func (c *connectionTCP) StartReader() bool {
	var msgByte []byte
	// 将连接内的数据流全部读取出来
	for {
		b := make([]byte, 512)
		if read, err := c.conn.Read(b); err != nil {
			return false
		} else {
			msgByte = append(msgByte, b[:read]...)
			if read < len(b) {
				break
			}
		}
	}

	// 将所有的内容分割成不同的消息，处理粘包
	msgData := defaultServer.DataPacket.UnPack(msgByte)
	if msgData == nil {
		return false
	}

	for _, data := range msgData {
		// 封装请求数据传入处理函数
		req := &request{conn: c, msg: data}
		if defaultServer.AppConf.WorkerPoolSize > 0 {
			GetInstanceMsgHandler().SendMsgToTaskQueue(req)
		} else {
			go GetInstanceMsgHandler().DoMsgHandler(req)
		}
	}
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

	GetInstanceBroadcastManager().GetGlobalBroadcastGroup().Append(c.GetConnId())
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
