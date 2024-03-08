package network

import (
	"context"
	"github.com/451008604/nets/iface"
	"io"
	"net"
	"sync"
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
	c.exitCtx, c.exitCtxCancel = context.WithCancel(context.Background())
	c.msgBuffChan = make(chan []byte, defaultServer.AppConf.MaxMsgChanLen)
	c.property = sync.Map{}
	return c
}

func (c *connectionTCP) StartReader() {
	// 获取客户端的消息头信息
	headData := make([]byte, defaultServer.DataPacket.GetHeadLen())
	if _, err := io.ReadFull(c.conn, headData); err != nil {
		GetInstanceConnManager().Remove(c)
		return
	}
	// 通过消息头获取dataLen和Id
	msgData := defaultServer.DataPacket.UnPack(headData)
	if msgData == nil {
		GetInstanceConnManager().Remove(c)
		return
	}
	// 通过消息头获取消息body
	if msgData.GetDataLen() > 0 {
		msgData.SetData(make([]byte, msgData.GetDataLen()))
		if _, err := io.ReadFull(c.conn, msgData.GetData()); err != nil {
			GetInstanceConnManager().Remove(c)
			return
		}
	}

	// 封装请求数据传入处理函数
	req := &request{conn: c, msg: msgData}
	if defaultServer.AppConf.WorkerPoolSize > 0 {
		GetInstanceMsgHandler().SendMsgToTaskQueue(req)
	} else {
		go GetInstanceMsgHandler().DoMsgHandler(req)
	}
}

func (c *connectionTCP) StartWriter(data []byte) {
	if _, err := c.conn.Write(data); err != nil {
		GetInstanceConnManager().Remove(c)
		return
	}
}

func (c *connectionTCP) Start(readerHandler func(), writerHandler func(data []byte)) {
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
