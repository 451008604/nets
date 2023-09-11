package api

import (
	"github.com/451008604/socketServerFrame/iface"
	"github.com/451008604/socketServerFrame/logic"
	"github.com/451008604/socketServerFrame/modules"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
)

func ItemCombineHandler(c iface.IConnection, message proto.Message) {
	p := logic.GetPlayer(c)
	req := message.(*pb.ItemCombineReq)
	res := &pb.ItemCombineRes{Result: proto.Uint32(modules.ErrSuccess)}

	if p.CheckCellIsOK(req.GetOriginIdx(), req.GetTargetIdx()) {
		res.Result = proto.Uint32(0)
		c.SendMsg(pb.MsgID_ItemCombine_Res, res)
		return
	}

}

func ItemProduceHandler(c iface.IConnection, message proto.Message) {

}
