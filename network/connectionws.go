package network

import (
	"errors"
	"github.com/451008604/socketServerFrame/config"
	"github.com/451008604/socketServerFrame/iface"
	"github.com/451008604/socketServerFrame/logs"
	"github.com/gorilla/websocket"
	"io"
	"sync"
)

type ConnectionWS struct {
	Connection
	conn *websocket.Conn
}

func NewConnectionWS(server iface.IServer, conn *websocket.Conn) *ConnectionWS {
	c := &ConnectionWS{}
	c.Server = server
	c.conn = conn
	c.ConnID = int(server.GetConnMgr().NewConnID())
	c.isClosed = false
	c.MsgHandler = GetInstanceMsgHandler()
	c.ExitBuffChan = make(chan bool, 1)
	c.msgBuffChan = make(chan []byte, config.GetGlobalObject().MaxMsgChanLen)
	c.property = make(map[string]interface{})
	c.propertyLock = sync.RWMutex{}
	return c
}

func (c *ConnectionWS) StartReader() {
	defer c.Stop()

	for {
		msgType, msgByte, err := c.conn.ReadMessage()
		if err != nil || msgType != websocket.BinaryMessage {
			if !errors.As(err, &io.ErrUnexpectedEOF) {
				logs.PrintLogErr(err)
			}
			return
		}

		msgData := c.Server.DataPacket().Unpack(msgByte)
		if msgData == nil {
			return
		}
		if msgData.GetDataLen() > 0 {
			msgData.SetData(msgByte[c.Server.DataPacket().GetHeadLen() : c.Server.DataPacket().GetHeadLen()+int(msgData.GetDataLen())])
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

func (c *ConnectionWS) StartWriter(data []byte) {
	err := c.conn.WriteMessage(websocket.BinaryMessage, data)
	logs.PrintLogErr(err, string(data))
}

func (c *ConnectionWS) Start(writerHandler func(data []byte)) {
	// 开启用于读的goroutine
	go c.StartReader()
	// 注册用于写的writerHandler
	c.Connection.Start(writerHandler)
}

func (c *ConnectionWS) Stop() {
	_ = c.conn.Close()

	c.Connection.Stop()
}

func (c *ConnectionWS) RemoteAddrStr() string {
	return c.conn.RemoteAddr().String()
}