package api

import (
	"github.com/451008604/socketServerFrame/iface"
	"google.golang.org/protobuf/proto"
)

func PingHandler(con iface.IConnection, message proto.Message) {
	// ping := message.(*pb.Ping)
	// // 更新返回时间戳
	// ping.TimeStamp = time.Now().UnixMicro() - ping.GetTimeStamp()
	//
	// // 发送给客户端的数据
	// con.SendMsg(pb.MessageID_PING, ProtocolToByte(ping))
}
