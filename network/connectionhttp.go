package network

import (
	"encoding/json"
	"github.com/451008604/nets/iface"
	"google.golang.org/protobuf/proto"
	"io"
	"net/http"
)

type connectionHTTP struct {
	connection
	writer http.ResponseWriter
	reader *http.Request
}

func NewConnectionHTTP(server iface.IServer, writer http.ResponseWriter, reader *http.Request) iface.IConnection {
	c := &connectionHTTP{}
	c.server = server
	c.writer = writer
	c.reader = reader
	c.isClosed = false
	c.msgBuffChan = make(chan []byte, defaultServer.AppConf.MaxMsgChanLen)
	c.property = NewConcurrentStringer[iface.IConnProperty, any]()
	c.taskQueue = make(chan func(), defaultServer.AppConf.WorkerTaskMaxLen)
	return c
}

type httpData struct {
	MsgID uint16 // 消息ID
	Data  string // 消息内容
}

func (c *connectionHTTP) StartReader() bool {
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
