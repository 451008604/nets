package api

import (
	"github.com/451008604/socketServerFrame/common"
	"github.com/451008604/socketServerFrame/dao/redis"
	"github.com/451008604/socketServerFrame/dao/sql"
	"github.com/451008604/socketServerFrame/dao/sqlmodel"
	"github.com/451008604/socketServerFrame/iface"
	"github.com/451008604/socketServerFrame/logic"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"strings"
)

func LoginHandler(c iface.IConnection, message proto.Message) {
	res := &pb.PlayerLoginResponse{
		Result:  proto.Int32(common.ErrSuccess),
		ReqData: message.(*pb.PlayerLoginRequest),
	}
	var (
		register = uint32(0)
		account  = &sqlmodel.HouseAccount{}
		user     = &sqlmodel.HouseUser{}
		err      error
	)

	switch res.ReqData.GetLoginType() {
	case common.LoginTypeQuick:
		if len(res.ReqData.GetAccount()) < 3 || len(res.ReqData.GetAccount()) > 80 {
			res.Result = proto.Int32(common.ErrAccountLengthErr)
			c.SendMsg(pb.MSG_ID_ID_S2C_PLAYER_LOGIN, res)
			return
		}
		if !strings.HasPrefix(res.ReqData.GetAccount(), res.ReqData.GetLoginType()) {
			res.ReqData.Account = proto.String(res.ReqData.GetLoginType() + "-" + res.ReqData.GetAccount())
		}
		// 查询账号信息 or 注册新账号
		register, account, user, err = sql.SQL.GetAccountInfo(res.ReqData.GetAccount(), res.ReqData.GetPassWord())
		if err != nil {
			if register == 0 {
				res.Result = proto.Int32(common.ErrPlayerInfoFetchFailed)
			} else {
				res.Result = proto.Int32(common.ErrRegisterFailed)
			}
			c.SendMsg(pb.MSG_ID_ID_S2C_PLAYER_LOGIN, res)
			return
		}

	default:
		res.Result = proto.Int32(common.ErrLoginTypeIllegal)
		c.SendMsg(pb.MSG_ID_ID_S2C_PLAYER_LOGIN, res)
		return
	}
	random, _ := uuid.NewRandom()
	res.Register = proto.Uint32(register)
	res.UserUniID = proto.Int64(user.UniID)
	res.Account = proto.String(account.Account)
	res.Password = proto.String(account.Password)
	res.RegisterTime = proto.Uint32(uint32(user.RegisterTime))
	res.RandomSeed = proto.Uint32(random.ID())

	// 初始化玩家数据
	player := logic.GetPlayer(c)
	player.InitializationSaveData()
	player.SetPlayerData(user)
	player.Random = common.NewRandom(int(random.ID()))
	// 读取缓存数据覆盖初始化数据
	if redis.Redis.GetPlayerInfo(user.UniID, player.Data) != nil {
		res.Result = proto.Int32(common.ErrPlayerInfoFetchFailed)
		c.SendMsg(pb.MSG_ID_ID_S2C_PLAYER_LOGIN, res)
		return
	}

	c.SendMsg(pb.MSG_ID_ID_S2C_PLAYER_LOGIN, res)

	// 推送玩家信息
	c.SendMsg(pb.MSG_ID_ID_S2C_PLAYER_DATA, player.Data)
}
