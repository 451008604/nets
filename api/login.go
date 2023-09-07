package api

import (
	"github.com/451008604/socketServerFrame/iface"
	"github.com/451008604/socketServerFrame/logic"
	"github.com/451008604/socketServerFrame/modules"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
)

func LoginHandler(conn iface.IConnection, message proto.Message) {
	msgReq := message.(*pb.PlayerLoginReq)
	msgRes := &pb.PlayerLoginRes{
		Result:      proto.Int32(modules.Success),
		Account:     proto.String(msgReq.GetAccount()),
		PassWord:    proto.String(msgReq.GetPassWord()),
		ChannelType: proto.Int32(msgReq.GetChannelType()),
	}

	switch msgReq.GetLoginType() {
	case modules.LoginTypeQuick:
		if len(msgReq.GetAccount()) < modules.UsernameLenMin || len(msgReq.GetAccount()) > modules.UsernameLenMax {
			msgRes.Result = proto.Int32(modules.AccountLengthErr)
			conn.SendMsg(pb.MsgID_PlayerLogin_Res, msgRes)
			return
		}
	}
	conn.SetPlayer(logic.Player{})
	println(conn.GetPlayer().(*logic.Player).Data.GetAccountData())
	conn.GetPlayer().(*logic.Player).Data.AccountData = &pb.PBAccountData{}
	println(conn.GetPlayer().(*logic.Player).Data.GetAccountData())

	conn.SendMsg(pb.MsgID_PlayerLogin_Res, msgRes)
}
