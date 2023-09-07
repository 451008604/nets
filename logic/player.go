package logic

import (
	"github.com/451008604/socketServerFrame/iface"
	pb "github.com/451008604/socketServerFrame/proto/bin"
)

type Player struct {
	conn iface.IConnection
	data pb.PBPlayerData
}
