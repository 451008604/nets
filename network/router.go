package network

import (
	"github.com/451008604/socketServerFrame/iface"
	"google.golang.org/protobuf/proto"
)

// router基类
type BaseRouter struct {
	Message proto.Message
	Handler iface.IReceiveMsgHandler
}

func (b *BaseRouter) SetMsg(message proto.Message) {
	b.Message = message
}

func (b *BaseRouter) GetMsg() proto.Message {
	return b.Message
}

func (b *BaseRouter) SetHandler(req iface.IReceiveMsgHandler) {
	b.Handler = req
}

func (b *BaseRouter) RunHandler(request iface.IRequest) {
	b.Handler(request.GetConnection(), b.Message)
}
