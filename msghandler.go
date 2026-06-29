package nets

import (
	"log/slog"
	"runtime/debug"
	"sync"
)

type IFilter func(conn IConnection, msg IMessage) bool
type IErrCapture func(conn IConnection, recover any)

type MsgHandler struct {
	apis       map[int32]*BaseRouter // Router Table / 路由表
	filter     IFilter               // Message Filter / 消息过滤器
	errCapture IErrCapture           // Error Capture / 错误捕获器
}

var instanceMsgHandler *MsgHandler
var instanceMsgHandlerOnce = sync.Once{}

// Message Handler / 消息处理器
func GetInstanceMsgHandler() *MsgHandler {
	instanceMsgHandlerOnce.Do(func() {
		instanceMsgHandler = &MsgHandler{
			apis: make(map[int32]*BaseRouter),
		}
	})
	return instanceMsgHandler
}

func (m *MsgHandler) AddRouter(msgId int32, msgTemplate INewMsgStructTemplate, msgHandler IReceiveMsgHandler) {
	if msgTemplate == nil || msgHandler == nil {
		slog.Error("router has nil template or handler", "msgId", msgId)
		return
	}
	if _, ok := m.apis[msgId]; ok {
		slog.Warn("msgId is duplicate", "msgId", msgId)
		return
	}
	m.apis[msgId] = &BaseRouter{}
	m.apis[msgId].SetMsg(msgTemplate)
	m.apis[msgId].SetHandler(msgHandler)
}

func (m *MsgHandler) GetRouter(msgId int32) (*BaseRouter, bool) {
	router, ok := m.apis[msgId]
	return router, ok
}

func (m *MsgHandler) SetFilter(fun IFilter) {
	m.filter = fun
}

func (m *MsgHandler) GetFilter(conn IConnection, msg IMessage) bool {
	if conn == nil || m.filter == nil {
		return true
	}
	defer m.GetErrCapture(conn)
	return m.filter(conn, msg)
}

func (m *MsgHandler) SetErrCapture(fun IErrCapture) {
	m.errCapture = fun
}

func (m *MsgHandler) GetErrCapture(conn IConnection) {
	if r1 := recover(); r1 != nil {
		slog.Error("panic recovered", "panic", r1, "stack", string(debug.Stack()))
		if m.errCapture != nil && conn != nil {
			defer func() {
				if r2 := recover(); r2 != nil {
					slog.Error("errCapture panic", "panic", r2, "stack", string(debug.Stack()))
				}
			}()
			m.errCapture(conn, r1)
		}
	}
}
