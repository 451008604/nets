package network

import (
	"github.com/451008604/nets/iface"
	"google.golang.org/protobuf/proto"
)

type baseRouter struct {
	template iface.INewMsgStructTemplate
	handler  iface.IReceiveMsgHandler
}

func (b *baseRouter) SetMsg(msgTemplate iface.INewMsgStructTemplate) {
	b.template = msgTemplate
}

func (b *baseRouter) GetNewMsg() proto.Message {
	return b.template()
}

func (b *baseRouter) SetHandler(msgHandler iface.IReceiveMsgHandler) {
	b.handler = msgHandler
}

func (b *baseRouter) RunHandler(conn iface.IConnection, message proto.Message) {
	b.handler(conn, message)
}
