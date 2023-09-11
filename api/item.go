package api

import (
	"github.com/451008604/socketServerFrame/config"
	"github.com/451008604/socketServerFrame/iface"
	"github.com/451008604/socketServerFrame/logic"
	"github.com/451008604/socketServerFrame/modules"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
)

func ItemProduceHandler(c iface.IConnection, message proto.Message) {

}

func ItemCombineHandler(c iface.IConnection, message proto.Message) {
	p := logic.GetPlayer(c)
	req := message.(*pb.ItemCombineReq)
	res := &pb.ItemCombineRes{Result: proto.Uint32(modules.ErrSuccess)}

	originItem, targetItem := p.GetItemDataByIdx(req.GetOriginIdx()), p.GetItemDataByIdx(req.GetTargetIdx())
	if originItem == nil || targetItem == nil || originItem.GetItemID() != targetItem.GetItemID() {
		res.Result = proto.Uint32(0)
		c.SendMsg(pb.MsgID_ItemCombine_Res, res)
		return
	}

	originJson, targetJson := config.GetItemConfig(int(originItem.GetItemID())), config.GetItemConfig(int(originItem.GetItemID()))
	if (originJson.IsProduceIntervalMinus() && originItem.GetCDEndTime() > 0) || (targetJson.IsProduceIntervalMinus() && targetItem.GetCDEndTime() > 0) {
		res.Result = proto.Uint32(0)
		c.SendMsg(pb.MsgID_ItemCombine_Res, res)
		return
	}

	if targetJson.SameNextID == 0 {
		res.Result = proto.Uint32(0)
		c.SendMsg(pb.MsgID_ItemCombine_Res, res)
		return
	}

	// 消除旧格子
	p.ClearCellItem(req.GetOriginIdx())

	newItem := p.NewItem(targetJson.SameNextID)
	p.SetCell(req.GetTargetIdx(), newItem)
	// 剩余生产次数继承
	if targetJson.InheritType == 2 {
		// 单周期道具的继承, 必须是合成双方都生产过道具, 否则不继承
		if originItem.GetProduceRemainNum() < uint32(originJson.ProduceNum) && targetItem.GetProduceRemainNum() < uint32(targetJson.ProduceNum) {
			newItem.ProduceRemainNum = proto.Uint32(originItem.GetProduceRemainNum() + targetItem.GetProduceRemainNum())
		}
	}

}
