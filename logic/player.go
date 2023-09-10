package logic

import (
	"github.com/451008604/socketServerFrame/dao/sqlmodel"
	"github.com/451008604/socketServerFrame/iface"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
)

type Player struct {
	Conn iface.IConnection
	Data *pb.PBPlayerData
}

func (p *Player) Initialization() *pb.PBPlayerData {
	p.Data = &pb.PBPlayerData{
		CommonData: &pb.PBCommonData{},
	}

	return p.Data
}

func (p *Player) SetPlayerData(userID int32, user *sqlmodel.HouseUser) {
	p.Data.CommonData.UserID = proto.Uint32(uint32(userID))
	p.Data.CommonData.NickName = proto.String(user.Nickname)
	p.Data.CommonData.HeadImage = proto.String(user.HeadImage)
	p.Data.CommonData.RegisterTime = proto.Uint32(uint32(user.RegisterTime))
}
