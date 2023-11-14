package network

import (
	"context"
	"errors"
	"fmt"
	"github.com/451008604/nets/config"
	"github.com/451008604/nets/iface"
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
	c.ConnID = GetInstanceConnManager().NewConnID()
	c.isClosed = false
	c.exitCtx, c.exitCtxCancel = context.WithCancel(context.Background())
	c.msgBuffChan = make(chan []byte, config.GetGlobalObject().MaxMsgChanLen)
	c.property = make(map[string]any)
	c.propertyLock = sync.RWMutex{}
	c.broadcastGroupByID = sync.Map{}
	c.broadcastGroupCh = make(chan iface.IBroadcastData, 1000)
	return c
}

func (c *ConnectionWS) StartReader() {
	msgType, msgByte, err := c.conn.ReadMessage()
	if err != nil || msgType != websocket.BinaryMessage {
		if !errors.As(err, &io.ErrUnexpectedEOF) {
			fmt.Printf("ws reader err %v\n", err)
		}
		GetInstanceConnManager().Remove(c)
		return
	}

	packet := c.Server.DataPacket()
	msgData := packet.Unpack(msgByte)
	if msgData == nil {
		GetInstanceConnManager().Remove(c)
		return
	}
	if msgData.GetDataLen() > 0 {
		msgData.SetData(msgByte[packet.GetHeadLen() : packet.GetHeadLen()+int(msgData.GetDataLen())])
	}

	// 封装请求数据传入处理函数
	req := &Request{conn: c, msg: msgData}
	if config.GetGlobalObject().WorkerPoolSize > 0 {
		GetInstanceMsgHandler().SendMsgToTaskQueue(req)
	} else {
		go GetInstanceMsgHandler().DoMsgHandler(req)
	}
}

func (c *ConnectionWS) StartWriter(data []byte) {
	if err := c.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		fmt.Printf("ws writer err %v data %v\n", err, data)
	}
}

func (c *ConnectionWS) Start(readerHandler func(), writerHandler func(data []byte)) {
	c.JoinBroadcastGroup(c, GetInsBroadcastManager().GetGlobalBroadcast())
	c.Connection.Start(readerHandler, writerHandler)
}

func (c *ConnectionWS) Stop() {
	if c.isClosed {
		return
	}
	c.Connection.Stop()
	_ = c.conn.Close()
	c.ExitAllBroadcastGroup()
}

func (c *ConnectionWS) RemoteAddrStr() string {
	return c.conn.RemoteAddr().String()
}
