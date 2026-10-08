package nets

import (
	"context"
	"google.golang.org/protobuf/proto"
	"io"
	"net"
	"net/http"
	"sync"
)

type connectionHTTP struct {
	*ConnectionBase
	writer        http.ResponseWriter
	reader        *http.Request
	writeMu       sync.Mutex
	headerWritten bool
	stopManager   func() bool
}

func NewConnectionHTTP(server IServer, writer http.ResponseWriter, reader *http.Request) IConnection {
	c := &connectionHTTP{ConnectionBase: newConnectionBase(server, reader.Context(), false), writer: writer, reader: reader}
	// HTTP writes synchronously and never consumes a long-connection queue.
	c.msgBuffChan = nil
	c.ConnectionBase.conn = c
	c.stopManager = context.AfterFunc(c.manager.ctx, c.connCtxCancel)
	return c
}
func (c *connectionHTTP) GetNetConn() net.Conn { return nil }

func (c *connectionHTTP) Close() { c.stopManager(); c.ConnectionBase.Close() }
func (c *connectionHTTP) Open()  { c.StartReader() }
func (c *connectionHTTP) DoTask(task func()) bool {
	if task == nil || c.IsClose() {
		return false
	}
	defer c.manager.messages.GetErrCapture(c)
	task()
	return true
}

var (
	ConnPropertyHttpAuthorization = "HttpAuthorization"
	ConnPropertyHttpReader        = "HttpReader"
	ConnPropertyHttpWriter        = "HttpWriter"
)

func (c *connectionHTTP) StartReader() bool {
	defer c.Close()
	if c.IsClose() {
		return false
	}
	c.SetProperty(ConnPropertyHttpAuthorization, c.reader.Header.Get(ConnPropertyHttpAuthorization))
	data, err := io.ReadAll(c.reader.Body)
	if err != nil {
		return false
	}
	msg := GetMessage()
	defer PutMessage(msg)
	if err := c.ByteToProtocol(data, msg); err != nil || msg.GetMsgId() == 0 {
		// Only an explicitly registered raw route may receive a non-envelope body.
		if _, ok := c.manager.messages.GetRouter(0); !ok {
			return false
		}
		msg.Reset()
		msg.SetData(data)
		c.SetProperty(ConnPropertyHttpReader, c.reader)
		c.SetProperty(ConnPropertyHttpWriter, c.writer)
	}
	if uint(len(msg.GetData())) > c.manager.config.AppConf.MaxPackSize {
		return false
	}
	defer func() { c.Close(); c.manager.connections.GetConnClosed(c) }()
	defer c.manager.messages.GetErrCapture(c)
	c.manager.connections.GetConnOpened(c)
	readerTaskHandler(c, msg)
	return true
}
func (c *connectionHTTP) write(data []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	if c.IsClose() {
		return ErrConnectionClosed
	}
	if !c.headerWritten {
		c.writer.WriteHeader(http.StatusOK)
		c.headerWritten = true
	}
	n, err := c.writer.Write(data)
	if err == nil && n != len(data) {
		err = io.ErrShortWrite
	}
	return err
}
func (c *connectionHTTP) StartWriter(data []byte) bool { return c.write(data) == nil }
func (c *connectionHTTP) RemoteAddrStr() string        { return c.reader.RemoteAddr }
func (c *connectionHTTP) SendMsgWithContext(ctx context.Context, id int32, msg proto.Message) error {
	if c.IsClose() {
		return ErrConnectionClosed
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	var data []byte
	var err error
	if id == 0 {
		data, err = c.marshal(msg)
	} else {
		data, err = c.pack(id, msg)
	}
	if err != nil {
		return err
	}
	if id == 0 && uint(len(data)) > c.manager.config.AppConf.MaxPackSize {
		return ErrInvalidMessage
	}
	return c.write(data)
}
