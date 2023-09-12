package api

import (
	"github.com/451008604/socketServerFrame/iface"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
	"time"
)

func HeartBeatHandler(con iface.IConnection, _ proto.Message) {
	// 发送给客户端的数据
	con.SendMsg(pb.MSG_ID_ID_S2C_HEARTBEAT, &pb.HeartbeatResponse{ServerTime: proto.Uint32(uint32(time.Now().Unix()))})
}
