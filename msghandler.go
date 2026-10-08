package nets

import (
	"google.golang.org/protobuf/proto"
	"log/slog"
	"runtime/debug"
	"sync"
)

type IFilter func(IConnection, IMessage) bool
type IErrCapture func(IConnection, any)
type MsgHandler struct {
	mu         sync.RWMutex
	apis       map[int32]*BaseRouter
	filter     IFilter
	errCapture IErrCapture
	logger     *slog.Logger
}

func (m *MsgHandler) AddSessionRouter(id int32, template INewMsgStructTemplate, handler func(Session, proto.Message)) {
	if handler == nil {
		return
	}
	m.AddRouter(id, template, func(conn IConnection, msg proto.Message) { handler(conn, msg) })
}
func NewMsgHandler() *MsgHandler {
	return &MsgHandler{apis: make(map[int32]*BaseRouter), logger: slog.Default()}
}
func GetInstanceMsgHandler() *MsgHandler { return getDefaultServerManager().messages }
func (m *MsgHandler) AddRouter(id int32, template INewMsgStructTemplate, handler IReceiveMsgHandler) {
	if id < 0 || id > 65535 || template == nil || handler == nil {
		m.logger.Error("invalid router", "msgId", id)
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.apis[id]; ok {
		m.logger.Warn("duplicate router", "msgId", id)
		return
	}
	m.apis[id] = &BaseRouter{template: template, handler: handler}
}
func (m *MsgHandler) GetRouter(id int32) (*BaseRouter, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	router, ok := m.apis[id]
	return router, ok
}
func (m *MsgHandler) SetFilter(filter IFilter) { m.mu.Lock(); m.filter = filter; m.mu.Unlock() }
func (m *MsgHandler) GetFilter(conn IConnection, msg IMessage) bool {
	m.mu.RLock()
	filter := m.filter
	m.mu.RUnlock()
	if conn == nil || filter == nil {
		return true
	}
	defer m.GetErrCapture(conn)
	return filter(conn, msg)
}
func (m *MsgHandler) SetErrCapture(capture IErrCapture) {
	m.mu.Lock()
	m.errCapture = capture
	m.mu.Unlock()
}
func (m *MsgHandler) GetErrCapture(conn IConnection) {
	if recovered := recover(); recovered != nil {
		m.logger.Error("panic recovered", "panic", recovered, "stack", string(debug.Stack()))
		m.mu.RLock()
		capture := m.errCapture
		m.mu.RUnlock()
		if capture != nil && conn != nil {
			defer func() {
				if secondary := recover(); secondary != nil {
					m.logger.Error("errCapture panic", "panic", secondary)
				}
			}()
			capture(conn, recovered)
		}
	}
}
