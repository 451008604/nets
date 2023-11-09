package network

import (
	"github.com/451008604/nets/iface"
	"google.golang.org/protobuf/proto"
)

// router基类
type BaseRouter struct {
	Message iface.INewMsgStructTemplate
	Handler iface.IReceiveMsgHandler
}

func (b *BaseRouter) SetMsg(msgStructTemplate iface.INewMsgStructTemplate) {
	b.Message = msgStructTemplate
}

func (b *BaseRouter) GetNewMsg() proto.Message {
	return b.Message()
}

func (b *BaseRouter) SetHandler(req iface.IReceiveMsgHandler) {
	b.Handler = req
}

func (b *BaseRouter) RunHandler(request iface.IRequest, message proto.Message) {
	b.Handler(request.GetConnection(), message)
}
