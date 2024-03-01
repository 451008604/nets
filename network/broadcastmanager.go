package network

import (
	"github.com/451008604/nets/iface"
	"google.golang.org/protobuf/proto"
	"sync"
	"sync/atomic"
)

type broadcastManager struct {
	idFlag          int64
	notifyList      sync.Map
	globalBroadcast iface.IBroadcast
}

var instanceBroadcastManager *broadcastManager
var instanceBroadcastManagerOnce = sync.Once{}

// 全局唯一广播管理器
func GetInstanceBroadcastManager() iface.IBroadcastManager {
	instanceBroadcastManagerOnce.Do(func() {
		instanceBroadcastManager = &broadcastManager{
			idFlag:     1000000000,
			notifyList: sync.Map{},
		}
		instanceBroadcastManager.globalBroadcast = instanceBroadcastManager.NewBroadcastGroup()
	})
	return instanceBroadcastManager
}

func (n *broadcastManager) NewBroadcastGroup() iface.IBroadcast {
	atomic.AddInt64(&n.idFlag, 1)
	broadcast := &broadcastGroup{
		groupID:    n.idFlag,
		targetList: sync.Map{},
	}
	n.notifyList.Store(broadcast.GetGroupID(), broadcast)
	return broadcast
}

func (n *broadcastManager) GetGlobalBroadcast() iface.IBroadcast {
	return n.globalBroadcast
}

func (n *broadcastManager) GetBroadcast(groupID int64) (any, bool) {
	value, ok := n.notifyList.Load(groupID)
	return value.(iface.IBroadcast), ok
}

func (n *broadcastManager) DelBroadcastByID(groupID int64) {
	n.notifyList.Delete(groupID)
}

func (n *broadcastManager) SendBroadcastData(groupID int64, connID int, msgID int32, data proto.Message) {
	if broadcast, ok := n.GetBroadcast(groupID); ok {
		broadcast.(iface.IBroadcast).BroadcastAllTargets(connID, msgID, data)
	}
}
