package api

import (
	"github.com/451008604/socketServerFrame/iface"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
)

func LoginHandler(con iface.IConnection, message proto.Message) {
	login := message.(*pb.PlayerLoginReq)
	// logs.PrintLogInfo(login.String())

	con.SendMsg(pb.MsgID_PlayerLogin_Res, []byte("登录成功 "+login.GetAccount()+" "+login.GetPassWord()))
}
