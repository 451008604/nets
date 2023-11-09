package api

import (
	"github.com/451008604/nets/iface"
	pb "github.com/451008604/nets/proto/bin"
	"google.golang.org/protobuf/proto"
	"time"
)

func HeartBeatHandler(c iface.IConnection, _ proto.Message) {
	// 发送给客户端的数据
	c.SendMsg(pb.MSgID_Heartbeat_Res, &pb.HeartbeatResponse{ServerTime: uint32(time.Now().Unix())})
}
