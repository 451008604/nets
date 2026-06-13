package nets

import (
	"fmt"
	"runtime/debug"
	"sync"
)

type IFilter func(conn IConnection, msg IMessage) bool
type IErrCapture func(conn IConnection, panicInfo string)

type MsgHandler struct {
	apis       map[int32]*BaseRouter // Router Table / 路由表
	apisMutex  sync.RWMutex           // Map R/W Lock / 地图读写锁
	filter     IFilter                // Message Filter / 消息过滤器
	errCapture IErrCapture            // Error Capture / 错误捕获器
	fieldMutex sync.RWMutex           // Protects filter and errCapture / 保护 filter 和 errCapture
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
		fmt.Printf("router %v has nil template or handler\n", msgId)
		return
	}
	m.apisMutex.Lock()
	defer m.apisMutex.Unlock()
	if _, ok := m.apis[msgId]; ok {
		fmt.Printf("msgId is duplicate %v\n", msgId)
		return
	}
	m.apis[msgId] = &BaseRouter{}
	m.apis[msgId].SetMsg(msgTemplate)
	m.apis[msgId].SetHandler(msgHandler)
}

func (m *MsgHandler) GetRouter(msgId int32) (*BaseRouter, bool) {
	m.apisMutex.RLock()
	defer m.apisMutex.RUnlock()
	router, ok := m.apis[msgId]
	return router, ok
}

func (m *MsgHandler) GetApis() map[int32]*BaseRouter {
	m.apisMutex.RLock()
	defer m.apisMutex.RUnlock()
	// Return a shallow copy to prevent concurrent modification
	copy := make(map[int32]*BaseRouter, len(m.apis))
	for k, v := range m.apis {
		copy[k] = v
	}
	return copy
}

func (m *MsgHandler) SetFilter(fun IFilter) {
	m.fieldMutex.Lock()
	defer m.fieldMutex.Unlock()
	m.filter = fun
}

func (m *MsgHandler) GetFilter() IFilter {
	m.fieldMutex.RLock()
	defer m.fieldMutex.RUnlock()
	return m.filter
}

func (m *MsgHandler) SetErrCapture(fun IErrCapture) {
	m.fieldMutex.Lock()
	defer m.fieldMutex.Unlock()
	m.errCapture = fun
}

func (m *MsgHandler) GetErrCapture(conn IConnection) {
	if conn == nil {
		return
	}
	m.fieldMutex.RLock()
	capture := m.errCapture
	m.fieldMutex.RUnlock()
	if capture == nil {
		return
	}
	if r := recover(); r != nil {
		capture(conn, fmt.Sprintf("%v\n%s", r, debug.Stack()))
	}
}
