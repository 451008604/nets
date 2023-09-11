package api

import (
	"github.com/451008604/socketServerFrame/iface"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
)

func HeartBeatHandler(con iface.IConnection, _ proto.Message) {
	// 发送给客户端的数据
	con.SendMsg(pb.MsgID_HeartBeat_Res, &pb.HeartBeatRes{})
}
