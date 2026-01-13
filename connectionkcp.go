package nets

import (
	"context"
	"net"
	"sync"
)

type connectionKCP struct {
	*ConnectionBase
	conn net.Conn
}

func NewConnectionKCP(server *serverKCP, conn net.Conn) IConnection {
	c := &connectionKCP{
		ConnectionBase: &ConnectionBase{
			server:        server,
			msgBuffChan:   make(chan []byte, defaultServer.AppConf.MaxMsgChanLen),
			taskQueue:     make(chan func(), defaultServer.AppConf.WorkerTaskMaxLen),
			property:      map[string]any{},
			propertyMutex: sync.RWMutex{},
		},
		conn: conn,
	}
	c.connId = c.RemoteAddrStr()
	c.exitCtx, c.exitCtxCancel = context.WithCancel(context.Background())
	c.ConnectionBase.conn = c
	return c
}

func (c *connectionKCP) StartReader() bool {
	// 获取消息头信息
	msgHead := make([]byte, defaultServer.DataPack.GetHeadLen())
	if read, err := c.conn.Read(msgHead); err != nil || read < defaultServer.DataPack.GetHeadLen() {
		return false
	}

	// 解析头信息
	msgData := defaultServer.DataPack.UnPack(msgHead)
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

func (c *connectionKCP) StartWriter(data []byte) bool {
	if _, err := c.conn.Write(data); err != nil {
		return false
	}
	return true
}

func (c *connectionKCP) Stop() bool {
	if !c.ConnectionBase.Stop() {
		return false
	}
	_ = c.conn.Close()
	return true
}

func (c *connectionKCP) RemoteAddrStr() string {
	return c.conn.RemoteAddr().String()
}
