package api

import (
	"github.com/451008604/socketServerFrame/common"
	"github.com/451008604/socketServerFrame/config"
	"github.com/451008604/socketServerFrame/iface"
	"github.com/451008604/socketServerFrame/logic"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
)

func ItemProduceHandler(c iface.IConnection, message proto.Message) {

}

func ItemCombineHandler(c iface.IConnection, message proto.Message) {
	p := logic.GetPlayer(c)
	req := message.(*pb.ItemCombineRequest)
	res := &pb.ItemCombineResponse{Result: proto.Uint32(common.ErrSuccess)}

	originItem, targetItem := p.GetItemDataByIdx(req.GetOriginIdx()), p.GetItemDataByIdx(req.GetTargetIdx())
	if originItem == nil || targetItem == nil || originItem.GetItemID() != targetItem.GetItemID() {
		res.Result = proto.Uint32(0)
		c.SendMsg(pb.MSG_ID_ID_S2C_ITEM_COMBINE, res)
		return
	}

	originJson, targetJson := config.GetItemConfig(int(originItem.GetItemID())), config.GetItemConfig(int(originItem.GetItemID()))
	if (originJson.IsProduceIntervalMinus() && originItem.GetCDEndTime() > 0) || (targetJson.IsProduceIntervalMinus() && targetItem.GetCDEndTime() > 0) {
		res.Result = proto.Uint32(0)
		c.SendMsg(pb.MSG_ID_ID_S2C_ITEM_COMBINE, res)
		return
	}

	if targetJson.SameNextID == 0 {
		res.Result = proto.Uint32(0)
		c.SendMsg(pb.MSG_ID_ID_S2C_ITEM_COMBINE, res)
		return
	}

	// 消除旧格子
	p.ClearCellItem(req.GetOriginIdx())
	// 填充新格子
	newItem := p.NewItem(targetJson.SameNextID)
	p.SetCell(req.GetTargetIdx(), newItem)

	// 额外掉落
	// rndIdx := int32(-1)
	// rndIdx = common.GetItemSpace(p.Data.GetItemSpaceData().GetItemData(), false, req.GetTargetIdx())
	// if rndIdx >= 0 || rndIdx < int32(p.GetItemSize()) {
	// 	bDrop := false
	// 	if _, ok := common.CombineRndLeastMap[p.DailyCombineTimes]; ok {
	// 		// 保底掉落
	// 		bDrop = true
	// 	} else {
	// 		// 随机决定是否掉落道具(万分比)
	// 		prob := mathmod.RandNum(0, 9999)
	// 		itemConfig := itemmod.GetItemJson(newItem.ID)
	// 		if itemConfig.ID > 0 && prob < int(itemConfig.CombineRndProbability) {
	// 			bDrop = true
	// 		}
	// 	}
	//
	// 	// 开始掉落
	// 	if bDrop {
	// 		itemConfig := itemmod.GetItemJson(newItem.ID)
	// 		if itemConfig.CombineRndDropId > 0 {
	// 			dropJson := itemmod.GetDropJson(itemConfig.CombineRndDropId)
	// 			if dropJson.ID > 0 {
	// 				newDropItem := itemmod.GetNewItem(dropJson.DropItemByRand())
	// 				p.SetItem(uint32(rndIdx), assistID, newDropItem)
	// 			}
	// 		}
	// 	}
	// }
}
