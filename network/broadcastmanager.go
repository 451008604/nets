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
	broadcastGroupByGroupId ConcurrentMap[Integer, iface.IMutexArray] // 广播组列表		[key: groupId, 	value: iface.IMutexArray(connIds)]
	broadcastGroupByConnId  ConcurrentMap[Integer, iface.IMutexArray] // 连接绑定的广播组	[key: connId, 	value: iface.IMutexArray(groupIds)]
	globalBroadcastGroup    iface.IMutexArray                         // 全局广播组
}

var instanceBroadcastManager *broadcastManager
var instanceBroadcastManagerOnce = sync.Once{}

// 广播管理器
func GetInstanceBroadcastManager() iface.IBroadcastManager {
	instanceBroadcastManagerOnce.Do(func() {
		instanceBroadcastManager = &broadcastManager{
			idFlag:                  1000000000,
			broadcastGroupByGroupId: NewConcurrentStringer[Integer, iface.IMutexArray](),
			broadcastGroupByConnId:  NewConcurrentStringer[Integer, iface.IMutexArray](),
		}
		instanceBroadcastManager.globalBroadcastGroup = instanceBroadcastManager.NewBroadcastGroup()
		go instanceBroadcastManager.autoClearEmptyBroadcastGroup()
	})
	return instanceBroadcastManager
}

func (n *broadcastManager) NewBroadcastGroup() iface.IMutexArray {
	atomic.AddUint32(&n.idFlag, 1)
	if store, ok := n.broadcastGroupByGroupId.Get(Integer(n.idFlag)); ok {
		return store
	}
	result := &broadcastGroup{}
	n.broadcastGroupByGroupId.Set(Integer(n.idFlag), result)
	return result
}

func (n *broadcastManager) GetGlobalBroadcastGroup() iface.IMutexArray {
	return n.globalBroadcastGroup
}

func (n *broadcastManager) JoinBroadcastGroup(groupId int, connId int) {
	if mutexArray, ok := n.broadcastGroupByGroupId.Get(Integer(groupId)); ok {
		mutexArray.Append(connId)
	} else {
		n.broadcastGroupByGroupId.Set(Integer(groupId), &broadcastGroup{arr: []int{connId}})
	}

	if mutexArray, ok := n.broadcastGroupByConnId.Get(Integer(connId)); ok {
		mutexArray.Append(groupId)
	} else {
		n.broadcastGroupByConnId.Set(Integer(connId), &broadcastGroup{arr: []int{groupId}})
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
	if value, ok := n.broadcastGroupByGroupId.Get(Integer(groupId)); ok {
		return value, ok
	} else {
		return nil, false
	}
}

func (n *broadcastManager) GetBroadcastGroupByConnId(connId int) (iface.IMutexArray, bool) {
	if value, ok := n.broadcastGroupByConnId.Get(Integer(connId)); ok {
		return value, ok
	} else {
		return nil, false
	}
}

func (n *broadcastManager) autoClearEmptyBroadcastGroup() {
	for range time.Tick(time.Minute) {
		for groupId, array := range n.broadcastGroupByGroupId.Items() {
			if len(array.GetArray()) == 0 {
				n.broadcastGroupByGroupId.Remove(groupId)
			}
		}

		for connId, array := range n.broadcastGroupByConnId.Items() {
			if len(array.GetArray()) == 0 {
				n.broadcastGroupByConnId.Remove(connId)
			}
		}
	}
}
