package network

import (
	"github.com/451008604/nets/iface"
	"sync"
	"sync/atomic"
)

type broadcastManager struct {
	idFlag               int64
	broadcastGroupList   sync.Map
	globalBroadcastGroup iface.IBroadcastGroup
}

var instanceBroadcastManager *broadcastManager
var instanceBroadcastManagerOnce = sync.Once{}

// 全局唯一广播管理器
func GetInstanceBroadcastManager() iface.IBroadcastManager {
	instanceBroadcastManagerOnce.Do(func() {
		instanceBroadcastManager = &broadcastManager{
			idFlag:             1000000000,
			broadcastGroupList: sync.Map{},
		}
		instanceBroadcastManager.globalBroadcastGroup = instanceBroadcastManager.NewBroadcastGroup()
	})
	return instanceBroadcastManager
}

func (n *broadcastManager) NewBroadcastGroup() iface.IBroadcastGroup {
	atomic.AddInt64(&n.idFlag, 1)
	broadcast := &broadcastGroup{
		groupId:    n.idFlag,
		targetList: sync.Map{},
	}
	n.broadcastGroupList.Store(broadcast.GetGroupId(), broadcast)
	return broadcast
}

func (n *broadcastManager) GetGlobalBroadcastGroup() iface.IBroadcastGroup {
	return n.globalBroadcastGroup
}

func (n *broadcastManager) GetBroadcastGroupById(groupId int64) (iface.IBroadcastGroup, bool) {
	value, ok := n.broadcastGroupList.Load(groupId)
	return value.(iface.IBroadcastGroup), ok
}

func (n *broadcastManager) DelBroadcastGroupById(groupId int64) {
	if broadcast, ok := n.GetBroadcastGroupById(groupId); ok {
		broadcast.ClearAllBroadcastTarget()
	}

	n.broadcastGroupList.Delete(groupId)
}
