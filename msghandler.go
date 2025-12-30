package main

import (
	"fmt"
	"runtime/debug"
	"sync"
)

type IFilter func(conn IConnection, msg IMessage) bool
type IErrCapture func(conn IConnection, msg IMessage, panicInfo string)

type MsgHandler struct {
	apis       map[int32]*BaseRouter // 路由表
	filter     IFilter               // 消息过滤器
	errCapture IErrCapture           // 错误捕获器
}

var instanceMsgHandler *MsgHandler
var instanceMsgHandlerOnce = sync.Once{}

// 消息处理器
func GetInstanceMsgHandler() *MsgHandler {
	instanceMsgHandlerOnce.Do(func() {
		instanceMsgHandler = &MsgHandler{
			apis: make(map[int32]*BaseRouter),
		}
	})
	return instanceMsgHandler
}

func (m *MsgHandler) AddRouter(msgId int32, msgTemplate INewMsgStructTemplate, msgHandler IReceiveMsgHandler) {
	if _, ok := m.apis[msgId]; ok {
		fmt.Printf("msgId is duplicate %v\n", msgId)
	}
	m.apis[msgId] = &BaseRouter{}
	m.apis[msgId].SetMsg(msgTemplate)
	m.apis[msgId].SetHandler(msgHandler)
}

func (m *MsgHandler) GetApis() map[int32]*BaseRouter {
	return m.apis
}

func (m *MsgHandler) SetFilter(fun IFilter) {
	m.filter = fun
}

func (m *MsgHandler) GetFilter() IFilter {
	return m.filter
}

func (m *MsgHandler) SetErrCapture(fun IErrCapture) {
	m.errCapture = fun
}

func (m *MsgHandler) GetErrCapture(conn IConnection, message IMessage) {
	if m.errCapture == nil {
		return
	}
	if r := recover(); r != nil {
		m.errCapture(conn, message, fmt.Sprintf("%v\n%s", r, debug.Stack()))
	}
}
