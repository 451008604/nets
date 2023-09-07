package api

import (
	"github.com/451008604/socketServerFrame/iface"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
)

func RegisterRouter(msgHandler iface.IMsgHandler) {
	// msgHandler.AddRouter(pb.MessageID_PING, func() proto.Message { return new(pb.Ping) }, PingHandler)
	msgHandler.AddRouter(pb.MsgID_PlayerLogin_Req, func() proto.Message { return new(pb.PlayerLoginReq) }, LoginHandler)
}
