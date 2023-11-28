package network

import (
	"github.com/451008604/nets/iface"
	"google.golang.org/protobuf/proto"
)

type baseRouter struct {
	message iface.INewMsgStructTemplate
	handler iface.IReceiveMsgHandler
}

func (b *baseRouter) SetMsg(msgStructTemplate iface.INewMsgStructTemplate) {
	b.message = msgStructTemplate
}

func (b *baseRouter) GetNewMsg() proto.Message {
	return b.message()
}

func (b *baseRouter) SetHandler(req iface.IReceiveMsgHandler) {
	b.handler = req
}

func (b *baseRouter) RunHandler(request iface.IRequest, message proto.Message) {
	b.handler(request.GetConnection(), message)
}
