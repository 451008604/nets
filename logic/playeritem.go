package logic

// 检查ItemIdx是否正常
func (p *Player) CheckItemIdxIsOK(itemIdx ...uint32) bool {
	temp := make(map[uint32]struct{})
	for _, idx := range itemIdx {
		if _, ok := temp[idx]; ok {
			return false
		}
		if idx < 0 || idx >= p.ItemSpaceSize() || idx >= p.ItemSpaceSize() {
			return false
		}
		temp[idx] = struct{}{}
	}
	return true
}

func (p *Player) CheckCellIsOK(itemIdx ...uint32) bool {
	if !p.CheckItemIdxIsOK(itemIdx...) {
		return false
	}
	for _, idx := range itemIdx {
		itemData := p.Data.GetItemSpaceData().GetItemData()[idx]
		if itemData.GetItemID() != 0 || itemData.GetUnLocked() == 0 {
			return false
		}
	}
	return true
}
