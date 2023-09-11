package config

import (
	"strconv"
	"strings"
)

type ItemJson struct {
	ID                     int    `json:"ID"`
	Name                   string `json:"Name"`
	Namekan                string `json:"Namekan"`
	GroupID                int    `json:"GroupID"`
	Icon                   string `json:"Icon"`
	UnlockDes              string `json:"UnlockDes"`
	UnlockDeskan           string `json:"UnlockDeskan"`
	CDDes                  string `json:"CDDes"`
	CDDeskan               string `json:"CDDeskan"`
	LockDes                string `json:"LockDes"`
	LockDeskan             string `json:"LockDeskan"`
	GroupInLine            int    `json:"GroupInLine"`
	Level                  int    `json:"level"`
	Type                   int    `json:"Type"`
	FunType                int    `json:"FunType"`
	FunValue               int    `json:"FunValue"`
	PropertyType           int    `json:"PropertyType"`
	PropertyValue          int    `json:"PropertyValue"`
	SellPrice              int    `json:"SellPrice"`
	SellTaskID             int    `json:"SellTaskID"`
	SameLastID             int    `json:"SameLastID"`
	SameNextID             int    `json:"SameNextID"`
	ProduceCost            int    `json:"ProduceCost"`
	ProduceID              int    `json:"ProduceID"`
	Lock                   bool   `json:"Lock"`
	UnLockType             string `json:"UnLockType"`
	UnLockPrice            int    `json:"UnLockPrice"`
	AutoCD                 bool   `json:"AutoCD"`
	ProduceOverID          int    `json:"ProduceOverID"`
	InheritType            int    `json:"InheritType"`
	ProduceTimes           int    `json:"ProduceTimes"`
	ProduceNum             int    `json:"ProduceNum"`
	ProduceInterval        string `json:"ProduceInterval"`
	CanAddTimes            int    `json:"CanAddTimes"`
	ResetPrice             int    `json:"ResetPrice"`
	Formula                string `json:"Formula"`
	ComposeFormula         string `json:"ComposeFormula"`
	Sellcheck              int    `json:"sellcheck"`
	DrawingsType           int    `json:"DrawingsType"`
	DrawingsID             int    `json:"DrawingsID"`
	Head                   string `json:"Head"`
	UnlockAwardId          int    `json:"UnlockAwardId"`
	UnlockAw               int    `json:"UnlockAw"`
	DailyOrderPrizeGold    string `json:"DailyOrderPrizeGold"`
	MarketPrice            string `json:"MarketPrice"`
	BubbleDiamond          string `json:"BubbleDiamond"`
	DailyOrderPrizeMagic   int    `json:"DailyOrderPrizeMagic"`
	SourceIDs              string `json:"SourceIDs"`
	DoubleClickToWarehouse int    `json:"DoubleClickToWarehouse"`
	SourceJump             int    `json:"SourceJump"`
	Sort                   int    `json:"Sort"`
	MoveToWarehouse        int    `json:"MoveToWarehouse"`
	ConfirmToUse           int    `json:"ConfirmToUse"`
	ADChangeID             int    `json:"ADChangeID"`
	BubbleADID             int    `json:"BubbleADID"`
	CombineRndProbability  int    `json:"CombineRndProbability"`
	CombineRndDropId       int    `json:"CombineRndDropId"`
	FirstDropId            int    `json:"FirstDropId"`
	AssistCanRecycle       int    `json:"AssistCanRecycle"`
	FunUseStr              string `json:"FunUseStr"`
	AutoProduceCD          int    `json:"AutoProduceCD"`
	AutoProduceDropID      int    `json:"AutoProduceDropID"`
	AutoProduceMaxLimit    int    `json:"AutoProduceMaxLimit"`
	PoolID                 int    `json:"PoolID"`
	DailyOrderWeight       int    `json:"DailyOrderWeight"`
	ClickSoundID           int    `json:"ClickSoundID"`
	EquivalentOutput       string `json:"EquivalentOutput"`
	BubbleProbability      int    `json:"BubbleProbability"`
	ImgRightUP             string `json:"ImgRightUP"`
	CompleteEffect         string `json:"CompleteEffect"`

	serverProduceInterval []int64 // 预解析字段, 产出周期间隔（单位 秒,负数代表预置CD）
	// MarketPriceList       []MarketPriceTemplate // 服务端自定义字段
}
type MarketPriceTemplate struct {
	Level    uint32
	Diamonds float32
}

var itemConfig = map[int]ItemJson{}
var itemConfigByGroupId = map[int][]ItemJson{}

// 加载道具配置
func init() {
	getConfigDataToBytes(jsonsPath, "Items.json", &itemConfig)
	for _, json := range itemConfig {
		// 缓存按组Id归类的道具列表
		if _, ok := itemConfigByGroupId[json.GroupID]; !ok {
			itemConfigByGroupId[json.GroupID] = []ItemJson{}
		}
		itemConfigByGroupId[json.GroupID] = append(itemConfigByGroupId[json.GroupID], json)

		// 预解析生产周期
		strSplit := strings.Split(json.ProduceInterval, ",")
		for _, strInterval := range strSplit {
			interval, _ := strconv.ParseInt(strInterval, 10, 64)
			json.serverProduceInterval = append(json.serverProduceInterval, interval)
		}

	}
}

func GetItemConfig(itemID int) ItemJson {
	return itemConfig[itemID]
}

func GetItemConfigByGroupID(groupID int) []ItemJson {
	return itemConfigByGroupId[groupID]
}

// 判断预置CD是否为负数
func (p *ItemJson) IsProduceIntervalMinus() bool {
	if len(p.serverProduceInterval) < 1 {
		return false
	}
	// 判断是否无限周期
	return p.serverProduceInterval[0] < 0
}
