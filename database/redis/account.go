package redis

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"google.golang.org/protobuf/proto"
	"io"
	"time"
)

type AccountInfo struct {
	TableName string
}

func NewAccountInfo() *AccountInfo {
	return &AccountInfo{
		TableName: "account:",
	}
}

func (a *AccountInfo) GetPBAccountData() *pb.PBAccountData {
	return &pb.PBAccountData{}
}

func (r *Module) GetAccountInfo(account, password string) (register int32, accountInfo *pb.PBAccountData, err error) {
	// TODO 暂时使用redis存储account，后续改为sql查询后存入redis
	accountData := r.account.GetPBAccountData()
	bytes, _ := r.Client.Get(r.Ctx, r.account.TableName+account).Bytes()
	_ = json.Unmarshal(bytes, accountData)
	// 新用户注册
	if accountData.GetUserID() == 0 {
		accountInfo, err = r.CreateNewAccount(account, password)
		return 1, accountInfo, err
	}

	return 0, accountData, nil
}

func (r *Module) CreateNewAccount(account, password string) (*pb.PBAccountData, error) {
	accountData := r.account.GetPBAccountData()
	newID, err := uuid.NewRandom()
	if err != nil {
		return accountData, err
	}
	encryption := md5.New()
	_, _ = io.WriteString(encryption, password)
	accountData.UserID = proto.Uint32(newID.ID())
	accountData.Account = proto.String(account)
	accountData.PassWord = proto.String(fmt.Sprintf("%x", encryption.Sum(nil)))
	accountData.NickName = proto.String(fmt.Sprintf("玩家_%v", newID.ID()))
	accountData.HeadImage = proto.String("1")
	accountData.PlayerLevel = proto.Int32(0)
	accountData.CreateTime = proto.Uint32(uint32(time.Now().Unix()))

	marshal, _ := json.Marshal(accountData)
	_, err = r.Client.Set(r.Ctx, r.account.TableName+account, marshal, redis.KeepTTL).Result()
	if err != nil {
		return accountData, err
	}
	return accountData, nil
}
