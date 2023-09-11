package api

import (
	"github.com/451008604/socketServerFrame/iface"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
)

func RegisterRouter(msgHandler iface.IMsgHandler) {
	msgHandler.AddRouter(pb.MsgID_HeartBeat_Req, func() proto.Message { return new(pb.HeartBeatReq) }, HeartBeatHandler)
	msgHandler.AddRouter(pb.MsgID_PlayerLogin_Req, func() proto.Message { return new(pb.PlayerLoginReq) }, LoginHandler)
	msgHandler.AddRouter(pb.MsgID_ItemCombine_Req, func() proto.Message { return new(pb.ItemCombineReq) }, ItemCombineHandler)
	msgHandler.AddRouter(pb.MsgID_ItemProduce_Req, func() proto.Message { return new(pb.ItemProduceReq) }, ItemProduceHandler)
}
