package network

import (
	"context"
	"github.com/451008604/nets/iface"
	"net"
)

type connectionTCP struct {
	*connectionBase
	conn *net.TCPConn // 当前连接对象
}

func NewConnectionTCP(server iface.IServer, conn *net.TCPConn) iface.IConnection {
	c := &connectionTCP{
		connectionBase: &connectionBase{
			server:      server,
			connId:      GetInstanceConnManager().NewConnId(),
			msgBuffChan: make(chan []byte, defaultServer.AppConf.MaxMsgChanLen),
			taskQueue:   make(chan func(), defaultServer.AppConf.WorkerTaskMaxLen),
			property:    NewConcurrentMap[any](),
		},
		conn: conn,
	}
	c.exitCtx, c.exitCtxCancel = context.WithCancel(context.Background())
	c.connectionBase.conn = c
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
	c.DoTask(func() { readerTaskHandler(c, msgData) })
	return true
}

func (c *connectionTCP) StartWriter(data []byte) bool {
	if _, err := c.conn.Write(data); err != nil {
		return false
	}
	return true
}

func (c *connectionTCP) Stop() bool {
	if !c.connectionBase.Stop() {
		return false
	}
	_ = c.conn.Close()
	return true
}

func (c *connectionTCP) RemoteAddrStr() string {
	return c.conn.RemoteAddr().String()
}
