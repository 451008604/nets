package config

import (
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
)

// 新玩家配置
type cellJson struct {
	Index       int32
	ID          int32
	Lock        int32
	BoxType     int32
	UnlockLevel int32
}

// 新玩家默认道具表
var cellJsons = []cellJson{}

// 格子信息列表（按锁定级别）
var cellInfoByLockLevelList = make(map[int32][]cellJson)

// 新玩家该格子是否有蛛网
func IsCellLocked(idx uint32) bool {
	if idx >= uint32(len(cellJsons)) {
		return false
	}
	return cellJsons[idx].Lock == 1
}

// 加载玩家默认道具背包
func init() {
	// 初始化变量
	jsonMap := make(map[string]cellJson)
	getConfigDataToBytes(jsonsPath, "Cells.json", &jsonMap)

	cellJsons = make([]cellJson, len(jsonMap))
	cellInfoByLockLevelList = make(map[int32][]cellJson)
	// json文件中的index从1开始
	for _, cell := range jsonMap {
		cellJsons[cell.Index-1] = cell

		if _, ok := cellInfoByLockLevelList[cell.UnlockLevel]; !ok {
			cellInfoByLockLevelList[cell.UnlockLevel] = []cellJson{}
		}
		cellInfoByLockLevelList[cell.UnlockLevel] = append(cellInfoByLockLevelList[cell.UnlockLevel], cell)
	}
}

// 初始化玩家道具背包, json中的index从1开始
func InitItemBag(itemList []*pb.PBItemData) {
	for i := 0; i < len(itemList); i++ {
		itemList[i] = &pb.PBItemData{
			ItemID:           proto.Uint32(uint32(cellJsons[i].ID)),
			CDEndTime:        nil,
			ProduceRemainNum: nil,
			ProduceCycle:     nil,
		}
	}
}

// GetCellInfoByLockLevel 按锁定级别获取单元格信息
func GetCellInfoByLockLevel(lockLevel int32) (cells []cellJson) {
	return cellInfoByLockLevelList[lockLevel]
}
