package api

import (
	"github.com/451008604/socketServerFrame/iface"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
)

func RegisterRouter(msgHandler iface.IMsgHandler) {
	msgHandler.AddRouter(pb.MSgID_Heartbeat_Req, func() proto.Message { return new(pb.HeartbeatRequest) }, HeartBeatHandler)
	msgHandler.AddRouter(pb.MSgID_PlayerLogin_Req, func() proto.Message { return new(pb.PlayerLoginRequest) }, LoginHandler)
}

func RegisterRouterClient(msgHandler iface.IMsgHandler) {
	msgHandler.AddRouter(pb.MSgID_Heartbeat_Res, func() proto.Message { return new(pb.HeartbeatResponse) }, nil)
	msgHandler.AddRouter(pb.MSgID_PlayerLogin_Res, func() proto.Message { return new(pb.PlayerLoginResponse) }, nil)
}
