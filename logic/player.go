package logic

import (
	"github.com/451008604/socketServerFrame/common"
	"github.com/451008604/socketServerFrame/dao/sqlmodel"
	"github.com/451008604/socketServerFrame/iface"
	pb "github.com/451008604/socketServerFrame/proto/bin"
)

type Player struct {
	Conn   iface.IConnection
	Data   *pb.PBPlayerData
	Random *common.Random // 随机数工具
}

// 初始化玩家默认数据结构
func (p *Player) InitializationSaveData() *pb.PBPlayerData {
	// 初始化缓存变量
	p.Data = &pb.PBPlayerData{
		CommonData: &pb.PBCommonData{},
	}

	return p.Data
}

func (p *Player) SetPlayerData(user *sqlmodel.HouseUser) {
	p.Data.CommonData.UserUniID = user.UniID
	p.Data.CommonData.NickName = user.Nickname
	p.Data.CommonData.HeadImage = user.HeadImage
	p.Data.CommonData.RegisterTime = uint32(user.RegisterTime)
}
