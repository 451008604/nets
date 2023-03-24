package api

import (
	"github.com/451008604/socketServerFrame/iface"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
)

func LoginHandler(con iface.IConnection, message proto.Message) {
	login := message.(*pb.ReqLogin)
	login.GetUserName()
	login.GetPassWord()
	println(login.String())
}
