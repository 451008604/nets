package network

import (
	"github.com/451008604/nets/iface"
	"sync"
	"sync/atomic"
)

type broadcastManager struct {
	idFlag                  int64
	broadcastGroupByGroupId sync.Map // 广播组(根据组Id)
	broadcastGroupByConnId  sync.Map // 广播组(根据连接Id)
	globalBroadcastGroup    iface.IBroadcastGroup
}

var instanceBroadcastManager *broadcastManager
var instanceBroadcastManagerOnce = sync.Once{}

// 全局唯一广播管理器
func GetInstanceBroadcastManager() iface.IBroadcastManager {
	instanceBroadcastManagerOnce.Do(func() {
		instanceBroadcastManager = &broadcastManager{
			idFlag:                  1000000000,
			broadcastGroupByGroupId: sync.Map{},
			broadcastGroupByConnId:  sync.Map{},
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

	n.broadcastGroupByGroupId.Store(broadcast.GetGroupId(), broadcast)
	return broadcast
}

func (n *broadcastManager) GetGlobalBroadcastGroup() iface.IBroadcastGroup {
	return n.globalBroadcastGroup
}

func (n *broadcastManager) GetBroadcastGroupById(groupId int64) (iface.IBroadcastGroup, bool) {
	value, ok := n.broadcastGroupByGroupId.Load(groupId)
	return value.(iface.IBroadcastGroup), ok
}

func (n *broadcastManager) DelBroadcastGroupById(groupId int64) {
	n.broadcastGroupByGroupId.Delete(groupId)
}

func (n *broadcastManager) GetBroadcastGroupByConnId(connId int) ([]iface.IBroadcastGroup, bool) {
	value, ok := n.broadcastGroupByConnId.Load(connId)
	return value.([]iface.IBroadcastGroup), ok
}

func (n *broadcastManager) SetBroadcastGroupByConnId(connId int, broadcastGroup iface.IBroadcastGroup) {
	store, loaded := n.broadcastGroupByConnId.LoadOrStore(connId, []iface.IBroadcastGroup{broadcastGroup})
	if loaded {
		temp := store.([]iface.IBroadcastGroup)
		temp = append(temp, broadcastGroup)
		n.broadcastGroupByConnId.Store(connId, temp)
	}
}

func (n *broadcastManager) DelBroadcastGroupByConnId(connId int, broadcastGroup iface.IBroadcastGroup) {
	n.broadcastGroupByConnId.Delete(connId)
	groups, b := n.GetBroadcastGroupByConnId(connId)
	if b {
		for i, group := range groups {
			if group.GetGroupId() == broadcastGroup.GetGroupId() {
				groups = append(groups[:i], groups[i+1:]...)
				n.broadcastGroupByConnId.Store(connId, groups)
				break
			}
		}
	}
}
