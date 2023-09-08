package api

import (
	"github.com/451008604/socketServerFrame/iface"
	"github.com/451008604/socketServerFrame/modules"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
	"strings"
)

func LoginHandler(c iface.IConnection, message proto.Message) {
	res := &pb.PlayerLoginRes{
		Result:  proto.Int32(modules.ErrSuccess),
		ReqData: message.(*pb.PlayerLoginReq),
	}
	var (
		register = int32(0)
		account  = &pb.PBAccountData{}
		err      error
	)

	switch res.ReqData.GetLoginType() {
	case modules.LoginTypeQuick:
		if len(res.ReqData.GetAccount()) < 3 || len(res.ReqData.GetAccount()) > 80 {
			res.Result = proto.Int32(modules.ErrAccountLengthErr)
			c.SendMsg(pb.MsgID_PlayerLogin_Res, res)
			return
		}
		if !strings.HasPrefix(res.ReqData.GetAccount(), res.ReqData.GetLoginType()) {
			res.ReqData.Account = proto.String(res.ReqData.GetLoginType() + "-" + res.ReqData.GetAccount())
		}
		register, account, err = modules.Module.Redis().GetAccountInfo(res.ReqData.GetAccount(), res.ReqData.GetPassWord())
		if err != nil {
			res.Result = proto.Int32(modules.ErrRegisterFailed)
			c.SendMsg(pb.MsgID_PlayerLogin_Res, res)
			return
		}
		res.Register = proto.Int32(register)

	default:
		res.Result = proto.Int32(modules.ErrLoginTypeIllegal)
		c.SendMsg(pb.MsgID_PlayerLogin_Res, res)
		return
	}

	playerInfo := modules.Module.Redis().GetPlayerInfo(account.GetUserID())
	if playerInfo == nil {
		res.Result = proto.Int32(modules.ErrPlayerInfoFetchFailed)
		c.SendMsg(pb.MsgID_PlayerLogin_Res, res)
		return
	}

	c.SendMsg(pb.MsgID_PlayerLogin_Res, res)

	// 推送玩家信息
	playerInfo.AccountData = account
	c.SendMsg(pb.MsgID_PlayerInfo_Notify, playerInfo)
}
