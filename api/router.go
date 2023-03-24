package api

import (
	"github.com/451008604/socketServerFrame/iface"
	pb "github.com/451008604/socketServerFrame/proto/bin"
)

func RegisterRouter(server iface.IServer) {
	server.AddRouter(pb.MessageID_PING, &pb.Ping{}, PingHandler)
	server.AddRouter(pb.MessageID_Login, &pb.ReqLogin{}, LoginHandler)
}
