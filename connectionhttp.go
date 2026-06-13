package nets

import (
	"context"
	"google.golang.org/protobuf/proto"
	"io"
	"net"
	"net/http"
	"sync/atomic"
)

type connectionHTTP struct {
	*ConnectionBase
	writer        http.ResponseWriter
	reader        *http.Request
	headerWritten int32
}

func NewConnectionHTTP(server IServer, writer http.ResponseWriter, reader *http.Request) IConnection {
	msgChanLen := defaultServer.AppConf.MaxMsgChanLen
	if msgChanLen < 0 {
		msgChanLen = 0
	}
	c := &connectionHTTP{
		ConnectionBase: &ConnectionBase{
			server:      server,
			connId:      GenerateConnID(),
			msgBuffChan: make(chan []byte, msgChanLen),
			property:    map[string]any{},
		},
		writer: writer,
		reader: reader,
	}
	c.connCtx, c.connCtxCancel = context.WithCancel(context.Background())
	c.ConnectionBase.conn = c
	return c
}

func (c *connectionHTTP) GetNetConn() net.Conn {
	return nil
}

var (
	ConnPropertyHttpAuthorization = "HttpAuthorization"
	ConnPropertyHttpReader        = "HttpReader"
	ConnPropertyHttpWriter        = "HttpWriter"
)

func (c *connectionHTTP) StartReader() bool {
	xToken := c.reader.Header.Get(ConnPropertyHttpAuthorization)
	c.SetProperty(ConnPropertyHttpAuthorization, xToken)

	// Parse Body Structure / 解析body结构
	data, err := io.ReadAll(c.reader.Body)
	if err != nil {
		return false
	}
	_ = c.reader.Body.Close()
	msgData := GetMessage()
	if err := c.ByteToProtocol(data, msgData); err != nil || msgData.GetMsgId() == 0 {
		msgData.SetData(data)
		c.SetProperty(ConnPropertyHttpReader, c.reader)
		c.SetProperty(ConnPropertyHttpWriter, c.writer)
	}

	defer func() {
		GetInstanceMsgHandler().GetErrCapture(c)
		PutMessage(msgData)
		GetInstanceConnManager().GetConnClosed(c)
	}()
	readerTaskHandler(c, msgData)
	return true
}

func (c *connectionHTTP) StartWriter(data []byte) bool {
	if atomic.CompareAndSwapInt32(&c.headerWritten, 0, 1) {
		c.writer.WriteHeader(http.StatusOK)
	}
	if _, err := c.writer.Write(data); err != nil {
		return false
	}
	return true
}

func (c *connectionHTTP) RemoteAddrStr() string {
	return c.reader.RemoteAddr
}

func (c *connectionHTTP) SendMsg(msgId int32, msgData proto.Message) {
	if c.IsClose() {
		return
	}

	// Send to Client / 发送给客户端
	var msgByte = c.ProtocolToByte(msgData)
	if msgId != 0 {
		packMsg := GetMessage()
		packMsg.Id = uint16(msgId)
		packMsg.Data = msgByte
		msgByte = defaultServer.DataPack.Pack(packMsg)
		PutMessage(packMsg)
	}
	c.StartWriter(msgByte)
}
