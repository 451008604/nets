package logic

import (
	"errors"
	"github.com/451008604/socketServerFrame/logs"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
)

// 检查道具索引是否合法
func (p *Player) CheckItemIdxIsOK(itemIdx ...uint32) bool {
	temp := make(map[uint32]struct{})
	for _, idx := range itemIdx {
		if _, ok := temp[idx]; ok {
			logs.PrintLogErr(errors.New(""), "CheckItemIdxIsOK")
			return false
		}
		if idx < 0 || idx >= p.ItemSpaceSize() {
			logs.PrintLogErr(errors.New(""), "CheckItemIdxIsOK")
			return false
		}
		temp[idx] = struct{}{}
	}
	return true
}

// 根据索引获取道具数据
func (p *Player) GetItemDataByIdx(itemIdx uint32) *pb.PBItemData {
	if !p.CheckItemIdxIsOK(itemIdx) {
		return nil
	}
	itemData := p.Data.GetItemSpaceData().GetItemData()[itemIdx]
	return itemData
}

func (p *Player) NewItem(itemID int) *pb.PBItemData {
	return &pb.PBItemData{
		ItemID: proto.Uint32(uint32(itemID)),
	}
}

func (p *Player) SetCell(itemIdx uint32, item *pb.PBItemData) {
	if !p.CheckItemIdxIsOK(itemIdx) {
		return
	}
	p.Data.GetItemSpaceData().GetItemData()[itemIdx] = item
}

func (p *Player) ClearCellItem(itemIdx uint32) {
	if !p.CheckItemIdxIsOK(itemIdx) {
		return
	}
	p.Data.GetItemSpaceData().GetItemData()[itemIdx] = &pb.PBItemData{}
}
