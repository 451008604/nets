package network

import (
	"fmt"
	"github.com/451008604/nets/iface"
	"sync"
)

type msgHandler struct {
	apis       map[int32]iface.IRouter // 路由表
	filter     iface.IFilter           // 消息过滤器
	errCapture iface.IErrCapture       // 错误捕获器
}

var instanceMsgHandler iface.IMsgHandler
var instanceMsgHandlerOnce = sync.Once{}

// 消息处理器
func GetInstanceMsgHandler() iface.IMsgHandler {
	instanceMsgHandlerOnce.Do(func() {
		instanceMsgHandler = &msgHandler{
			apis: make(map[int32]iface.IRouter),
		}
	})
	return instanceMsgHandler
}

func (m *msgHandler) AddRouter(msgId int32, msg iface.INewMsgStructTemplate, handler iface.IReceiveMsgHandler) {
	if _, ok := m.apis[msgId]; ok {
		fmt.Printf("msgId is duplicate %v\n", msgId)
	}
	m.apis[msgId] = &baseRouter{}
	m.apis[msgId].SetMsg(msg)
	m.apis[msgId].SetHandler(handler)
}

func (m *msgHandler) GetApis() map[int32]iface.IRouter {
	return m.apis
}

func (m *msgHandler) SetFilter(fun iface.IFilter) {
	m.filter = fun
}

func (m *msgHandler) GetFilter() iface.IFilter {
	return m.filter
}

func (m *msgHandler) SetErrCapture(fun iface.IErrCapture) {
	m.errCapture = fun
}

func (m *msgHandler) GetErrCapture(conn iface.IConnection) {
	if m.errCapture == nil {
		return
	}
	if r := recover(); r != nil {
		m.errCapture(conn, r)
	}
}
