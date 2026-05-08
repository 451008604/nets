package nets

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type connectionWS struct {
	*ConnectionBase
	conn *websocket.Conn
}

func NewConnectionWS(server IServer, conn *websocket.Conn) IConnection {
	c := &connectionWS{
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
	c.connCtx, c.connCtxCancel = context.WithCancel(context.Background())
	c.ConnectionBase.conn = c
	return c
}

func (c *connectionWS) GetNetConn() net.Conn {
	return c.conn.NetConn()
}

func (c *connectionWS) StartReader() bool {
	msgType, msgByte, err := c.conn.ReadMessage()
	if err != nil || msgType != websocket.BinaryMessage {
		return false
	}

	for len(msgByte) > 0 {
		msgData := defaultServer.DataPack.UnPack(msgByte)
		if msgData == nil {
			return false
		}
		msgByte = msgByte[int(msgData.GetDataLen())+defaultServer.DataPack.GetHeadLen():]

		c.DoTask(func() {
			readerTaskHandler(c, msgData)
			PutMessage(msgData)
		})
	}
	return true
}

func (c *connectionWS) StartWriter(data []byte) bool {
	if err := c.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
		return false
	}
	return true
}

func (c *connectionWS) RemoteAddrStr() string {
	return c.conn.RemoteAddr().String()
}
