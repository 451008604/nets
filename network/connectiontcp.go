package network

import (
	"github.com/451008604/socketServerFrame/config"
	"github.com/451008604/socketServerFrame/iface"
	"github.com/451008604/socketServerFrame/logs"
	"io"
	"net"
	"sync"
)

type ConnectionTCP struct {
	Connection
	conn *net.TCPConn // 当前连接对象
}

// 新建连接
func NewConnectionTCP(server iface.IServer, conn *net.TCPConn, msgHandler iface.IMsgHandler) *ConnectionTCP {
	c := &ConnectionTCP{}
	c.Server = server
	c.conn = conn
	c.ConnID = int(server.GetConnMgr().NewConnID())
	c.isClosed = false
	c.MsgHandler = msgHandler
	c.ExitBuffChan = make(chan bool, 1)
	c.msgBuffChan = make(chan []byte, config.GetGlobalObject().MaxMsgChanLen)
	c.property = make(map[string]interface{})
	c.propertyLock = sync.RWMutex{}
	return c
}

func (c *ConnectionTCP) StartReader() {
	defer c.Stop()

	for {
		// 获取客户端的消息头信息
		headData := make([]byte, c.Server.DataPacket().GetHeadLen())
		if _, err := io.ReadFull(c.conn, headData); err != nil {
			if err != io.EOF {
				logs.PrintLogErr(err)
			}
			return
		}
		// 通过消息头获取dataLen和Id
		msgData := c.Server.DataPacket().Unpack(headData)
		if msgData == nil {
			return
		}
		// 通过消息头获取消息body
		if msgData.GetDataLen() > 0 {
			msgData.SetData(make([]byte, msgData.GetDataLen()))
			if _, err := io.ReadFull(c.conn, msgData.GetData()); logs.PrintLogErr(err) {
				return
			}
		}

		// 封装请求数据传入处理函数
		req := &Request{conn: c, msg: msgData}
		if config.GetGlobalObject().WorkerPoolSize > 0 {
			c.MsgHandler.SendMsgToTaskQueue(req)
		} else {
			go c.MsgHandler.DoMsgHandler(req)
		}
	}
}

func (c *ConnectionTCP) StartWriter() {
	for data := range c.msgBuffChan {
		_, err := c.conn.Write(data)
		logs.PrintLogErr(err, string(data))
	}
}

func (c *ConnectionTCP) Start() {
	// 开启用于读的goroutine
	go c.StartReader()
	// 开启用于写的goroutine
	go c.StartWriter()

	c.Connection.Start()
}

func (c *ConnectionTCP) Stop() {
	_ = c.conn.Close()

	c.Connection.Stop()
}

// 获取客户端地址信息
func (c *ConnectionTCP) RemoteAddrStr() string {
	return c.conn.RemoteAddr().String()
}
