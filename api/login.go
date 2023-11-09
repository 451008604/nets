package api

import (
	"github.com/451008604/nets/common"
	"github.com/451008604/nets/dao/redis"
	"github.com/451008604/nets/dao/sql"
	"github.com/451008604/nets/dao/sqlmodel"
	"github.com/451008604/nets/iface"
	"github.com/451008604/nets/logic"
	pb "github.com/451008604/nets/proto/bin"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"strings"
)

func LoginHandler(c iface.IConnection, message proto.Message) {
	res := &pb.PlayerLoginResponse{
		Result:  common.ErrSuccess,
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
			res.Result = common.ErrAccountLengthErr
			c.SendMsg(pb.MSgID_PlayerLogin_Res, res)
			return
		}
		if !strings.HasPrefix(res.ReqData.GetAccount(), res.ReqData.GetLoginType()) {
			res.ReqData.Account = res.ReqData.GetLoginType() + "-" + res.ReqData.GetAccount()
		}
		// 查询账号信息 or 注册新账号
		register, account, user, err = sql.SQL.GetAccountInfo(res.ReqData.GetAccount(), res.ReqData.GetPassWord())
		if err != nil {
			if register == 0 {
				res.Result = common.ErrPlayerInfoFetchFailed
			} else {
				res.Result = common.ErrRegisterFailed
			}
			c.SendMsg(pb.MSgID_PlayerLogin_Res, res)
			return
		}

	default:
		res.Result = common.ErrLoginTypeIllegal
		c.SendMsg(pb.MSgID_PlayerLogin_Res, res)
		return
	}
	random, _ := uuid.NewRandom()
	res.Register = register
	res.UserUniID = user.UniID
	res.Account = account.Account
	res.Password = account.Password
	res.RegisterTime = uint32(user.RegisterTime)
	res.RandomSeed = random.ID()

	// 初始化玩家数据
	player := logic.GetPlayer(c)
	player.InitializationSaveData()
	player.SetPlayerData(user)
	player.Random = common.NewRandom(int(random.ID()))
	// 读取缓存数据覆盖初始化数据
	if redis.Redis.GetPlayerInfo(user.UniID, player.Data) != nil {
		res.Result = common.ErrPlayerInfoFetchFailed
		c.SendMsg(pb.MSgID_PlayerLogin_Res, res)
		return
	}

	c.SendMsg(pb.MSgID_PlayerLogin_Res, res)
}
