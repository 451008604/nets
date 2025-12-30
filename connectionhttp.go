package main

import (
	"context"
	"encoding/json"
	"google.golang.org/protobuf/proto"
	"io"
	"net/http"
)

type connectionHTTP struct {
	*ConnectionBase
	writer http.ResponseWriter
	reader *http.Request
}

func NewConnectionHTTP(server IServer, writer http.ResponseWriter, reader *http.Request) IConnection {
	c := &connectionHTTP{
		ConnectionBase: &ConnectionBase{
			server:      server,
			msgBuffChan: make(chan []byte, defaultServer.AppConf.MaxMsgChanLen),
			taskQueue:   make(chan func(), defaultServer.AppConf.WorkerTaskMaxLen),
			property:    NewConcurrentMap[any](),
		},
		writer: writer,
		reader: reader,
	}
	c.exitCtx, c.exitCtxCancel = context.WithCancel(context.Background())
	c.ConnectionBase.conn = c
	return c
}

var (
	ConnPropertyHttpAuthorization = "HttpAuthorization"
	ConnPropertyHttpReader        = "HttpReader"
	ConnPropertyHttpWriter        = "HttpWriter"
)

func (c *connectionHTTP) StartReader() bool {
	xToken := c.reader.Header.Get(ConnPropertyHttpAuthorization)
	ConnPropertySet(c, ConnPropertyHttpAuthorization, xToken)

	// 解析body结构
	data, _ := io.ReadAll(c.reader.Body)
	msgData := &Message{}
	if err := c.ByteToProtocol(data, msgData); err != nil || msgData.GetMsgId() == 0 {
		msgData.Data = string(data)
		ConnPropertySet(c, ConnPropertyHttpReader, c.reader)
		ConnPropertySet(c, ConnPropertyHttpWriter, c.writer)
	}

	readerTaskHandler(c, msgData)
	return true
}

func (c *connectionHTTP) StartWriter(data []byte) bool {
	c.writer.WriteHeader(http.StatusOK)
	if _, err := c.writer.Write(data); err != nil {
		return false
	}
	return true
}

func (c *connectionHTTP) RemoteAddrStr() string {
	return c.reader.RemoteAddr
}

func (c *connectionHTTP) SendMsg(msgId int32, msgData proto.Message) {
	if c.isClosed {
		return
	}
	res := &Message{
		Id:   uint16(msgId),
		Data: string(c.ProtocolToByte(msgData)),
	}
	bytes, _ := json.Marshal(res)
	// 发送给客户端
	c.StartWriter(bytes)
}
