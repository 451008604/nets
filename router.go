package nets

import (
	"google.golang.org/protobuf/proto"
)

type IReceiveMsgHandler func(conn IConnection, message proto.Message)
type INewMsgStructTemplate func() proto.Message

type BaseRouter struct {
	template INewMsgStructTemplate
	handler  IReceiveMsgHandler
}

func (b *BaseRouter) SetMsg(msgTemplate INewMsgStructTemplate) {
	b.template = msgTemplate
}

func (b *BaseRouter) GetNewMsg() proto.Message {
	return b.template()
}

func (b *BaseRouter) SetHandler(msgHandler IReceiveMsgHandler) {
	b.handler = msgHandler
}

func (b *BaseRouter) RunHandler(conn IConnection, message proto.Message) {
	b.handler(conn, message)
}
