package redis

import (
	"encoding/json"
	pb "github.com/451008604/socketServerFrame/proto/bin"
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

func (r *Module) GetPlayerInfo(userID uint32) *pb.PBPlayerData {
	playerData := r.player.GetPlayerData()
	bytes, _ := r.Client.Get(r.Ctx, r.player.TableName+strconv.Itoa(int(userID))).Bytes()
	_ = json.Unmarshal(bytes, playerData)

	if playerData.GetCommonData() == nil {
		return nil
	}

	return playerData
}
