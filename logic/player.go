package logic

import (
	"github.com/451008604/socketServerFrame/iface"
	pb "github.com/451008604/socketServerFrame/proto/bin"
)

type Player struct {
	Conn iface.IConnection
	Data *pb.PBPlayerData
}

func (p *Player) Initialization() *pb.PBPlayerData {
	p.Data = &pb.PBPlayerData{
		AccountData: &pb.PBAccountData{},
		CommonData:  &pb.PBCommonData{},
	}

	return p.Data
}
