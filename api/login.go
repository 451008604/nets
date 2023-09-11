package api

import (
	"github.com/451008604/socketServerFrame/dao/redis"
	"github.com/451008604/socketServerFrame/dao/sql"
	"github.com/451008604/socketServerFrame/dao/sqlmodel"
	"github.com/451008604/socketServerFrame/iface"
	"github.com/451008604/socketServerFrame/logic"
	"github.com/451008604/socketServerFrame/modules"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"github.com/google/uuid"
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
		account  = &sqlmodel.HouseAccount{}
		user     = &sqlmodel.HouseUser{}
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
		// 查询账号信息 or 注册新账号
		register, account, user, err = sql.SQL.GetAccountInfo(res.ReqData.GetAccount(), res.ReqData.GetPassWord())
		if err != nil {
			if register == 0 {
				res.Result = proto.Int32(modules.ErrPlayerInfoFetchFailed)
			} else {
				res.Result = proto.Int32(modules.ErrRegisterFailed)
			}
			c.SendMsg(pb.MsgID_PlayerLogin_Res, res)
			return
		}

	default:
		res.Result = proto.Int32(modules.ErrLoginTypeIllegal)
		c.SendMsg(pb.MsgID_PlayerLogin_Res, res)
		return
	}
	res.Register = proto.Int32(register)
	random, _ := uuid.NewRandom()
	res.RandomSeed = proto.Uint32(random.ID())

	// 初始化玩家数据
	player := logic.GetPlayer(c)
	player.InitializationSaveData()
	player.SetPlayerData(account.UserID, user)
	player.RandomSeed = random.ID()
	// 读取缓存数据覆盖初始化数据
	if redis.Redis.GetPlayerInfo(uint32(user.ID), player.Data) != nil {
		res.Result = proto.Int32(modules.ErrPlayerInfoFetchFailed)
		c.SendMsg(pb.MsgID_PlayerLogin_Res, res)
	}

	c.SendMsg(pb.MsgID_PlayerLogin_Res, res)

	// 推送玩家信息
	c.SendMsg(pb.MsgID_PlayerInfo_Notify, player.Data)
}
