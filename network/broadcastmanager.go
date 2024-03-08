package network

import (
	"github.com/451008604/nets/iface"
	"google.golang.org/protobuf/proto"
	"sync"
	"sync/atomic"
	"time"
)

type broadcastManager struct {
	idFlag                  uint32
	broadcastGroupByGroupId sync.Map          // 广播组列表		[key: groupId, 	value: iface.IMutexArray(connIds)]
	broadcastGroupByConnId  sync.Map          // 连接绑定的广播组	[key: connId, 	value: iface.IMutexArray(groupIds)]
	globalBroadcastGroup    iface.IMutexArray // 全局广播组
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
		go instanceBroadcastManager.autoClearEmptyBroadcastGroup()
	})
	return instanceBroadcastManager
}

func (n *broadcastManager) NewBroadcastGroup() iface.IMutexArray {
	atomic.AddUint32(&n.idFlag, 1)
	store, _ := n.broadcastGroupByGroupId.LoadOrStore(int(n.idFlag), &broadcastGroup{})
	return store.(iface.IMutexArray)
}

func (n *broadcastManager) GetGlobalBroadcastGroup() iface.IMutexArray {
	return n.globalBroadcastGroup
}

func (n *broadcastManager) JoinBroadcastGroup(groupId int, connId int) {
	if groupActual, groupLoaded := n.broadcastGroupByGroupId.LoadOrStore(groupId, connId); groupLoaded {
		groupActual.(iface.IMutexArray).Append(connId)
	}
	if connActual, connLoaded := n.broadcastGroupByConnId.LoadOrStore(connId, groupId); connLoaded {
		connActual.(iface.IMutexArray).Append(groupId)
	}
}

func (n *broadcastManager) ExitBroadcastGroup(groupId int, connId int) {
	if value, ok := n.GetBroadcastGroupByGroupId(groupId); ok {
		value.Remove(connId)
	}
	if value, ok := n.GetBroadcastGroupByConnId(connId); ok {
		value.Remove(groupId)
	}
}

func (n *broadcastManager) SendBroadcastAllTargets(groupId int, msgId int32, data proto.Message) {
	if load, ok := n.GetBroadcastGroupByGroupId(groupId); ok {
		array := load.GetArray()
		for _, connId := range array {
			if conn, b := GetInstanceConnManager().Get(connId); b {
				conn.SendMsg(msgId, data)
			}
		}
	}
}

func (n *broadcastManager) GetBroadcastGroupByGroupId(groupId int) (iface.IMutexArray, bool) {
	value, ok := n.broadcastGroupByGroupId.Load(groupId)
	return value.(iface.IMutexArray), ok
}

func (n *broadcastManager) GetBroadcastGroupByConnId(connId int) (iface.IMutexArray, bool) {
	value, ok := n.broadcastGroupByConnId.Load(connId)
	return value.(iface.IMutexArray), ok
}

func (n *broadcastManager) autoClearEmptyBroadcastGroup() {
	for range time.Tick(time.Minute) {
		n.broadcastGroupByGroupId.Range(func(key, value any) bool {
			if len(value.(*broadcastGroup).GetArray()) == 0 {
				n.broadcastGroupByGroupId.Delete(key)
			}
			return true
		})

		n.broadcastGroupByConnId.Range(func(key, value any) bool {
			if len(value.(*broadcastGroup).GetArray()) == 0 {
				n.broadcastGroupByConnId.Delete(key)
			}
			return true
		})
	}
}
