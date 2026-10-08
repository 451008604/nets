package nets

import (
	"google.golang.org/protobuf/proto"
	"sync"
)

type IReceiveMsgHandler func(IConnection, proto.Message)
type INewMsgStructTemplate func() proto.Message
type BaseRouter struct {
	mu       sync.RWMutex
	template INewMsgStructTemplate
	handler  IReceiveMsgHandler
}

func (r *BaseRouter) SetMsg(template INewMsgStructTemplate) {
	r.mu.Lock()
	r.template = template
	r.mu.Unlock()
}
func (r *BaseRouter) GetNewMsg() proto.Message {
	r.mu.RLock()
	template := r.template
	r.mu.RUnlock()
	return template()
}
func (r *BaseRouter) SetHandler(handler IReceiveMsgHandler) {
	r.mu.Lock()
	r.handler = handler
	r.mu.Unlock()
}
func (r *BaseRouter) RunHandler(conn IConnection, msg proto.Message) {
	r.mu.RLock()
	handler := r.handler
	r.mu.RUnlock()
	handler(conn, msg)
}
