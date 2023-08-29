package api

import (
	"github.com/451008604/socketServerFrame/iface"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
)

func RegisterRouter(server iface.IServer) {
	server.AddRouter(pb.MessageID_PING, func() proto.Message { return new(pb.Ping) }, PingHandler)
	server.AddRouter(pb.MessageID_Login, func() proto.Message { return new(pb.ReqLogin) }, LoginHandler)
}
