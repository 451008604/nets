package api

import (
	"github.com/451008604/socketServerFrame/iface"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
)

func RegisterRouter(msgHandler iface.IMsgHandler) {
	msgHandler.AddRouter(pb.MSG_ID_ID_C2S_HEARTBEAT, func() proto.Message { return new(pb.HeartbeatRequest) }, HeartBeatHandler)
	msgHandler.AddRouter(pb.MSG_ID_ID_C2S_PLAYER_LOGIN, func() proto.Message { return new(pb.PlayerLoginRequest) }, LoginHandler)
	msgHandler.AddRouter(pb.MSG_ID_ID_C2S_ITEM_COMBINE, func() proto.Message { return new(pb.ItemCombineRequest) }, ItemCombineHandler)
	msgHandler.AddRouter(pb.MSG_ID_ID_C2S_ITEM_PRODUCE, func() proto.Message { return new(pb.ItemProduceRequest) }, ItemProduceHandler)
}

func RegisterRouterClient(msgHandler iface.IMsgHandler) {
	msgHandler.AddRouter(pb.MSG_ID_ID_S2C_PLAYER_DATA, func() proto.Message { return new(pb.PBPlayerData) }, nil)
	msgHandler.AddRouter(pb.MSG_ID_ID_S2C_HEARTBEAT, func() proto.Message { return new(pb.HeartbeatResponse) }, nil)
	msgHandler.AddRouter(pb.MSG_ID_ID_S2C_PLAYER_LOGIN, func() proto.Message { return new(pb.PlayerLoginResponse) }, nil)
	msgHandler.AddRouter(pb.MSG_ID_ID_S2C_ITEM_COMBINE, func() proto.Message { return new(pb.ItemCombineResponse) }, nil)
	msgHandler.AddRouter(pb.MSG_ID_ID_S2C_ITEM_PRODUCE, func() proto.Message { return new(pb.ItemProduceResponse) }, nil)
}
