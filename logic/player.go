package logic

import (
	"github.com/451008604/socketServerFrame/dao/sqlmodel"
	"github.com/451008604/socketServerFrame/iface"
	"github.com/451008604/socketServerFrame/modules"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
)

type Player struct {
	Conn          iface.IConnection
	Data          *pb.PBPlayerData
	RandomSeed    uint32
	itemSpaceSize [2]uint32 // 道具空间大小 [宽, 高]
}

// 初始化玩家默认数据结构
func (p *Player) InitializationSaveData() *pb.PBPlayerData {
	// 初始化缓存变量
	p.Data = &pb.PBPlayerData{
		CommonData: &pb.PBCommonData{},
		ItemSpaceData: &pb.PBItemSpaceData{
			ItemData: make([]*pb.PBItemData, 0),
		},
	}
	for i := 0; i < int(p.ItemSpaceSize()); i++ {
		p.Data.GetItemSpaceData().ItemData = append(p.Data.GetItemSpaceData().ItemData, &pb.PBItemData{})
	}

	// 初始化临时变量
	p.itemSpaceSize = [2]uint32{modules.ItemSpaceWidth, modules.ItemSpaceHeight}

	return p.Data
}

func (p *Player) SetPlayerData(userID int32, user *sqlmodel.HouseUser) {
	p.Data.CommonData.UserID = proto.Uint32(uint32(userID))
	p.Data.CommonData.NickName = proto.String(user.Nickname)
	p.Data.CommonData.HeadImage = proto.String(user.HeadImage)
	p.Data.CommonData.RegisterTime = proto.Uint32(uint32(user.RegisterTime))
}

// 项目空间大小
func (p *Player) ItemSpaceSize() uint32 {
	return p.itemSpaceSize[0] * p.itemSpaceSize[1]
}
