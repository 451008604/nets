package config

import (
	"encoding/json"
	"github.com/451008604/socketServerFrame/logs"
	"github.com/451008604/socketServerFrame/modules"
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
}

// 加载道具配置
func init() {
	logs.PrintLogPanic(json.Unmarshal(getConfigDataToBytes(modules.ExportJsonPath, "Items.json"), &ItemJson{}))
}
