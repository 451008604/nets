package network

import (
	"github.com/451008604/nets/iface"
	"google.golang.org/protobuf/proto"
	"sync"
	"sync/atomic"
)

type BroadcastManager struct {
	idFlag          int64
	notifyList      sync.Map
	globalBroadcast iface.IBroadcast
}

var insBroadcastManager *BroadcastManager
var insBroadcastManagerOnce = sync.Once{}

func GetInsBroadcastManager() iface.IBroadcastManager {
	insBroadcastManagerOnce.Do(func() {
		insBroadcastManager = &BroadcastManager{
			idFlag:     1000000000,
			notifyList: sync.Map{},
		}
		insBroadcastManager.globalBroadcast = insBroadcastManager.NewBroadcastGroup()
	})
	return insBroadcastManager
}

// 新建广播组
func (n *BroadcastManager) NewBroadcastGroup() iface.IBroadcast {
	atomic.AddInt64(&n.idFlag, 1)
	broadcast := &BroadcastGroup{
		groupID:    n.idFlag,
		targetList: sync.Map{},
	}
	n.notifyList.Store(broadcast.GetGroupID(), broadcast)
	return broadcast
}

func (n *BroadcastManager) GetGlobalBroadcast() iface.IBroadcast {
	return n.globalBroadcast
}

func (n *BroadcastManager) GetBroadcast(groupID int64) (any, bool) {
	value, ok := n.notifyList.Load(groupID)
	return value.(iface.IBroadcast), ok
}

func (n *BroadcastManager) DelBroadcastByID(groupID int64) {
	n.notifyList.Delete(groupID)
}

func (n *BroadcastManager) SendBroadcastData(groupID int64, connID int, msgID int32, data proto.Message) {
	if broadcast, ok := n.GetBroadcast(groupID); ok {
		broadcast.(iface.IBroadcast).BroadcastAllTargets(connID, msgID, data)
	}
}
