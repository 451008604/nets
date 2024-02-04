package network

import (
	"context"
	"fmt"
	"github.com/451008604/nets/config"
	"github.com/451008604/nets/iface"
	"github.com/gorilla/websocket"
	"sync"
)

type connectionWS struct {
	connection
	conn *websocket.Conn
}

func NewConnectionWS(server iface.IServer, conn *websocket.Conn) iface.IConnection {
	c := &connectionWS{}
	c.server = server
	c.conn = conn
	c.connID = GetInstanceConnManager().NewConnID()
	c.isClosed = false
	c.exitCtx, c.exitCtxCancel = context.WithCancel(context.Background())
	c.msgBuffChan = make(chan []byte, config.GetServerConf().MaxMsgChanLen)
	c.property = sync.Map{}
	c.broadcastGroupByID = sync.Map{}
	c.broadcastGroupCh = make(chan iface.IBroadcastData, 1000)
	return c
}

func (c *connectionWS) StartReader() {
	msgType, msgByte, err := c.conn.ReadMessage()
	if err != nil || msgType != websocket.BinaryMessage {
		GetInstanceConnManager().Remove(c)
		return
	}

	packet := c.server.DataPacket()
	msgData := packet.UnPack(msgByte)
	if msgData == nil {
		GetInstanceConnManager().Remove(c)
		return
	}
	if msgData.GetDataLen() > 0 {
		msgData.SetData(msgByte[packet.GetHeadLen() : packet.GetHeadLen()+int(msgData.GetDataLen())])
	}

	// 封装请求数据传入处理函数
	req := &request{conn: c, msg: msgData}
	if config.GetServerConf().WorkerPoolSize > 0 {
		GetInstanceMsgHandler().SendMsgToTaskQueue(req)
	} else {
		go GetInstanceMsgHandler().DoMsgHandler(req)
	}
}

func (c *connectionWS) StartWriter(data []byte) {
	if err := c.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		fmt.Printf("ws writer err %v data %v\n", err, data)
	}
}

func (c *connectionWS) Start(readerHandler func(), writerHandler func(data []byte)) {
	defer GetInstanceConnManager().Remove(c)

	c.JoinBroadcastGroup(c, GetInsBroadcastManager().GetGlobalBroadcast())
	c.connection.Start(readerHandler, writerHandler)
}

func (c *connectionWS) Stop() {
	if c.isClosed {
		return
	}
	c.connection.Stop()
	_ = c.conn.Close()
	c.ExitAllBroadcastGroup()
}

func (c *connectionWS) RemoteAddrStr() string {
	return c.conn.RemoteAddr().String()
}
