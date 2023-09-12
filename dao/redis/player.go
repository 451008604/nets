package redis

import (
	"encoding/json"
	"github.com/451008604/socketServerFrame/logs"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"github.com/redis/go-redis/v9"
	"strconv"
)

type PlayerInfo struct {
	TableName string
}

func NewPlayerInfo() *PlayerInfo {
	return &PlayerInfo{
		TableName: "player:",
	}
}

// 读取玩家数据
func (r *Module) GetPlayerInfo(userID int64, initPlayerData *pb.PBPlayerData) error {
	bytes, _ := r.Client.Get(r.Ctx, r.player.TableName+strconv.Itoa(int(userID))).Bytes()
	if bytes != nil {
		err := json.Unmarshal(bytes, initPlayerData)
		logs.PrintLogErr(err)
		return err
	}
	return nil
}

// 保存玩家数据
func (r *Module) SetPlayerInfo(userID int64, playerInfo *pb.PBPlayerData) error {
	marshal, err := json.Marshal(playerInfo)
	if err != nil {
		return err
	}
	_, err = r.Client.Set(r.Ctx, r.player.TableName+strconv.Itoa(int(userID)), marshal, redis.KeepTTL).Result()
	if err != nil {
		return err
	}
	return nil
}
