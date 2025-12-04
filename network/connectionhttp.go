package network

import (
	"encoding/json"
	"github.com/451008604/nets/iface"
	"google.golang.org/protobuf/proto"
	"io"
	"net/http"
)

type connectionHTTP struct {
	*connection
	writer http.ResponseWriter
	reader *http.Request
}

func NewConnectionHTTP(server iface.IServer, writer http.ResponseWriter, reader *http.Request) iface.IConnection {
	c := &connectionHTTP{
		connection: &connection{
			server:      server,
			msgBuffChan: make(chan []byte, defaultServer.AppConf.MaxMsgChanLen),
			taskQueue:   make(chan func(), defaultServer.AppConf.WorkerTaskMaxLen),
			property:    NewConcurrentMap[any](),
		},
		writer: writer,
		reader: reader,
	}
	return c
}

type httpData struct {
	MsgID uint16 // 消息ID
	Data  string // 消息内容
}

func (c *connectionHTTP) StartReader() bool {
	xToken := c.reader.Header.Get("Authorization")
	ConnPropertySet(c.connection, "Authorization", xToken)

	// 解析body结构
	data, _ := io.ReadAll(c.reader.Body)
	req := &httpData{}
	if err := json.Unmarshal(data, req); err != nil {
		c.writer.WriteHeader(http.StatusBadRequest)
		_, _ = c.writer.Write([]byte("server is closed"))
		return false
	}

	// 写入message
	msgData := &message{}
	msgData.SetMsgId(req.MsgID)
	msgData.SetData([]byte(req.Data))
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
	res := &httpData{
		MsgID: uint16(msgId),
		Data:  string(c.ProtocolToByte(msgData)),
	}
	bytes, _ := json.Marshal(res)
	// 发送给客户端
	c.StartWriter(bytes)
}
