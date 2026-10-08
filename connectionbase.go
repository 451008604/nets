package nets

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"google.golang.org/protobuf/proto"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

var connIdSeed uint32

func GenerateConnID() string {
	now := uint64(time.Now().Unix())
	seq := uint64(atomic.AddUint32(&connIdSeed, 1))
	return strconv.FormatUint((now<<32)|seq, 16)
}

type ConnectionBase struct {
	manager       *ServerManager
	server        IServer
	conn          IConnection
	connId        string
	msgBuffChan   chan []byte
	property      map[string]any
	propertyMutex sync.RWMutex
	isClosed      int32
	connCtx       context.Context
	connCtxCancel context.CancelFunc
	limitingCount int64
	limitingTimer int64
	limitingMutex sync.Mutex
	taskMu        sync.Mutex
	tasksClosed   bool
	tasks         sync.WaitGroup
}

func newConnectionBase(server IServer, parent context.Context, queued bool) *ConnectionBase {
	r := managerForServer(server)
	ctx, cancel := context.WithCancel(parent)
	var messages chan []byte
	if queued {
		messages = make(chan []byte, r.config.AppConf.MaxMsgChanLen)
	}
	return &ConnectionBase{manager: r, server: server, connId: GenerateConnID(), msgBuffChan: messages, property: make(map[string]any), connCtx: ctx, connCtxCancel: cancel}
}
func (c *ConnectionBase) managerOwner() *ServerManager { return c.manager }
func (c *ConnectionBase) Open() {
	readerDone := make(chan struct{})
	readerStarted := false
	opened := false
	defer func() {
		c.Close()
		if readerStarted {
			<-readerDone
		}
		c.tasks.Wait()
		if opened {
			c.manager.connections.GetConnClosed(c.conn)
		}
		c.propertyMutex.Lock()
		clear(c.property)
		c.propertyMutex.Unlock()
	}()
	defer c.manager.messages.GetErrCapture(c.conn)
	if c.IsClose() {
		return
	}
	if secure, ok := c.conn.GetNetConn().(*tls.Conn); ok {
		// Handshake only after connection admission, outside the shared workers.
		// No business hook or reader runs before the peer has completed TLS.
		ctx, cancel := context.WithTimeout(c.ConnCtx(), time.Duration(c.manager.config.AppConf.ConnRWTimeOut)*time.Second)
		err := secure.HandshakeContext(ctx)
		cancel()
		if err != nil {
			c.manager.config.Logger.Debug("TLS handshake failed", "connId", c.connId, "err", err)
			return
		}
		if c.IsClose() {
			return
		}
	}
	opened = true
	c.manager.connections.GetConnOpened(c.conn)
	if c.IsClose() {
		return
	}
	readerStarted = true
	go func() { defer close(readerDone); c.readHandler() }()
	var writeDeadline time.Time
	timeout := time.Duration(c.manager.config.AppConf.ConnRWTimeOut) * time.Second
	for {
		select {
		case <-c.ConnCtx().Done():
			return
		case data := <-c.msgBuffChan:
			now := time.Now()
			if writeDeadline.IsZero() || now.Add(timeout/2).After(writeDeadline) {
				writeDeadline = now.Add(timeout)
				if conn := c.conn.GetNetConn(); conn != nil {
					if err := conn.SetWriteDeadline(writeDeadline); err != nil {
						return
					}
				}
			}
			if !c.conn.StartWriter(data) {
				return
			}
		}
	}
}
func (c *ConnectionBase) readHandler() {
	defer c.Close()
	defer c.manager.messages.GetErrCapture(c.conn)
	timeout := time.Duration(c.manager.config.AppConf.ConnRWTimeOut) * time.Second
	for {
		if c.IsClose() {
			return
		}
		if conn := c.conn.GetNetConn(); conn != nil {
			if err := conn.SetReadDeadline(time.Now().Add(timeout)); err != nil {
				return
			}
		}
		if !c.conn.StartReader() {
			return
		}
	}
}
func (c *ConnectionBase) ConnCtx() context.Context { return c.connCtx }
func (c *ConnectionBase) Close() {
	c.taskMu.Lock()
	c.tasksClosed = true
	c.taskMu.Unlock()
	if !atomic.CompareAndSwapInt32(&c.isClosed, 0, 1) {
		return
	}
	c.connCtxCancel()
	if c.conn != nil {
		if conn := c.conn.GetNetConn(); conn != nil {
			if secure, ok := conn.(*tls.Conn); ok {
				// Abort the underlying transport first: tls.Close's close_notify
				// write must not delay cancellation on a backpressured stream.
				_ = secure.NetConn().Close()
			}
			_ = conn.Close()
		}
	}
}

// DoTask never waits for queue space. Callers must handle false on rejection.
func (c *ConnectionBase) DoTask(task func()) bool { return c.doTask(task, false) }

// doReadTask preserves backpressure for protocol readers outside the worker pool.
func (c *ConnectionBase) doReadTask(task func()) bool { return c.doTask(task, true) }

func (c *ConnectionBase) doTask(task func(), block bool) bool {
	c.taskMu.Lock()
	if task == nil || c.tasksClosed || c.IsClose() {
		c.taskMu.Unlock()
		return false
	}
	// Close closes this admission gate before Open can wait for task completion.
	c.tasks.Add(1)
	c.taskMu.Unlock()
	pool := c.manager.WorkerPool()
	err := pool.submit(c.ConnCtx(), func() {
		defer c.tasks.Done()
		defer c.manager.messages.GetErrCapture(c.conn)
		task()
	}, pool.workerIndex(pool.HashWorkerId(c.connId)), block)
	if err != nil {
		c.tasks.Done()
	}
	return err == nil
}

func submitReadTask(conn IConnection, task func()) bool {
	if reader, ok := conn.(interface{ doReadTask(func()) bool }); ok {
		return reader.doReadTask(task)
	}
	return conn.DoTask(task)
}

func readerTaskHandler(c IConnection, msg IMessage) {
	r := managerForConnection(c)
	defer r.messages.GetErrCapture(c)
	if c.IsClose() || c.ConnCtx().Err() != nil {
		return
	}
	if c.FlowControl() || !r.messages.GetFilter(c, msg) {
		return
	}
	router, ok := r.messages.GetRouter(int32(msg.GetMsgId()))
	if !ok {
		return
	}
	var payload proto.Message
	if msg.GetMsgId() == 0 {
		raw, ok := msg.(*Message)
		if !ok {
			return
		}
		payload = raw
	} else {
		payload = router.GetNewMsg()
		if err := c.ByteToProtocol(msg.GetData(), payload); err != nil {
			r.config.Logger.Error("message parsing failed", "msgId", msg.GetMsgId(), "err", err)
			return
		}
	}
	router.RunHandler(c, payload)
}
func (c *ConnectionBase) GetConnId() string { return c.connId }
func (c *ConnectionBase) IsClose() bool {
	return atomic.LoadInt32(&c.isClosed) != 0 || c.ConnCtx().Err() != nil
}
func (c *ConnectionBase) GetProperty(key string) any {
	c.propertyMutex.RLock()
	defer c.propertyMutex.RUnlock()
	return c.property[key]
}
func (c *ConnectionBase) SetProperty(key string, value any) {
	c.propertyMutex.Lock()
	defer c.propertyMutex.Unlock()
	c.property[key] = value
}
func (c *ConnectionBase) RemoveProperty(key string) {
	c.propertyMutex.Lock()
	defer c.propertyMutex.Unlock()
	delete(c.property, key)
}

func (c *ConnectionBase) marshal(msg proto.Message) ([]byte, error) {
	if msg == nil {
		return nil, ErrInvalidMessage
	}
	if c.manager.config.AppConf.ProtocolIsJson {
		return json.Marshal(msg)
	}
	return proto.Marshal(msg)
}
func (c *ConnectionBase) pack(id int32, msg proto.Message) ([]byte, error) {
	if id < 0 || id > 65535 {
		return nil, fmt.Errorf("%w: message ID out of uint16 range", ErrInvalidMessage)
	}
	data, err := c.marshal(msg)
	if err != nil {
		return nil, err
	}
	if uint(len(data)) > c.manager.config.AppConf.MaxPackSize {
		return nil, fmt.Errorf("%w: payload too large", ErrInvalidMessage)
	}
	envelope := GetMessage()
	defer PutMessage(envelope)
	envelope.Id, envelope.Data = uint16(id), data
	if packer, ok := c.manager.config.DataPack.(ErrorDataPack); ok {
		return packer.PackWithError(envelope)
	}
	packed := c.manager.config.DataPack.Pack(envelope)
	if packed == nil {
		return nil, ErrInvalidMessage
	}
	return packed, nil
}

// SendMsgWithContext reports enqueue failures; successful enqueue does not confirm delivery.
func (c *ConnectionBase) SendMsgWithContext(ctx context.Context, id int32, msg proto.Message) error {
	if c.IsClose() {
		return ErrConnectionClosed
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	data, err := c.pack(id, msg)
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-c.ConnCtx().Done():
		return ErrConnectionClosed
	case c.msgBuffChan <- data:
		return nil
	}
}
func (c *ConnectionBase) SendMsg(id int32, msg proto.Message) {
	sender := ContextSender(c)
	if transport, ok := c.conn.(ContextSender); ok {
		sender = transport
	}
	if err := sender.SendMsgWithContext(c.ConnCtx(), id, msg); err != nil {
		c.manager.config.Logger.Warn("send message failed", "msgId", id, "err", err)
	}
}
func (c *ConnectionBase) FlowControl() bool {
	limit := int64(c.manager.config.AppConf.MaxFlowSecond)
	if limit == -1 {
		return false
	}
	now := time.Now().UnixMilli()
	c.limitingMutex.Lock()
	if c.limitingTimer == 0 || now-c.limitingTimer >= 1000 {
		c.limitingTimer = now
		c.limitingCount = 0
	}
	c.limitingCount++
	limited := c.limitingCount > limit
	c.limitingMutex.Unlock()
	if limited {
		c.manager.connections.ConnRateLimiting(c.conn)
	}
	return limited
}
func (c *ConnectionBase) RemoteAddrStr() string {
	if conn := c.conn.GetNetConn(); conn != nil {
		return conn.RemoteAddr().String()
	}
	return ""
}
func (c *ConnectionBase) ProtocolToByte(msg proto.Message) []byte {
	data, err := c.marshal(msg)
	if err != nil {
		c.manager.config.Logger.Error("message serialization failed", "err", err)
		return nil
	}
	return data
}
func (c *ConnectionBase) ByteToProtocol(data []byte, target proto.Message) error {
	if target == nil {
		return ErrInvalidMessage
	}
	if c.manager.config.AppConf.ProtocolIsJson {
		return json.Unmarshal(data, target)
	}
	if err := proto.Unmarshal(data, target); err != nil {
		return err
	}
	if msg, ok := target.(*Message); ok && msg.wireID > 65535 {
		return ErrInvalidMessage
	}
	return nil
}
