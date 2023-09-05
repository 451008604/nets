package api

import (
	"github.com/451008604/socketServerFrame/iface"
	"github.com/451008604/socketServerFrame/logs"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
)

func LoginHandler(con iface.IConnection, message proto.Message) {
	login := message.(*pb.ReqLogin)
	logs.PrintLogInfo(login.String())

	con.SendMsg(pb.MessageID_Login, []byte("登录成功 "+login.GetUserName()+" "+login.GetPassWord()))
}
