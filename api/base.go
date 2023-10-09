package api

import (
	"github.com/451008604/socketServerFrame/iface"
	"github.com/451008604/socketServerFrame/logic"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
)

func init() {
	// 注册路由
	RegisterRouter()
}

func MsgFilter(c iface.IRequest, req proto.Message) bool {
	// 未登录时不处理任何请求
	if c.GetMsgID() != pb.MSgID_PlayerLogin_Req && logic.GetPlayer(c.GetConnection()).Data == nil {
		return false
	}

	return true
}
