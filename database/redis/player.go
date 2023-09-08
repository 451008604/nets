package redis

import (
	"encoding/json"
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

func (p *PlayerInfo) GetPlayerData() *pb.PBPlayerData {
	return &pb.PBPlayerData{}
}

func (r *Module) GetPlayerInfo(userID uint32, initPlayerData *pb.PBPlayerData) (*pb.PBPlayerData, error) {
	bytes, _ := r.Client.Get(r.Ctx, r.player.TableName+strconv.Itoa(int(userID))).Bytes()
	_ = json.Unmarshal(bytes, initPlayerData)
	return initPlayerData, nil
}

func (r *Module) SetPlayerInfo(userID uint32, playerInfo *pb.PBPlayerData) error {
	marshal, _ := json.Marshal(playerInfo)
	_, err := r.Client.Set(r.Ctx, r.player.TableName+strconv.Itoa(int(userID)), marshal, redis.KeepTTL).Result()
	if err != nil {
		return err
	}
	return nil
}
