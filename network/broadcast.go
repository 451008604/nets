package network

import (
	"google.golang.org/protobuf/proto"
	"sync"
)

type broadcastGroup struct {
	groupId    int64
	targetList sync.Map
}

func (n *broadcastGroup) GetGroupId() int64 {
	return n.groupId
}

func (n *broadcastGroup) SetBroadcastTarget(connId int) {
	n.targetList.Store(connId, 1)
}

func (n *broadcastGroup) GetBroadcastTarget(connId int) bool {
	_, ok := n.targetList.Load(connId)
	return ok
}

func (n *broadcastGroup) DelBroadcastTarget(connId int) {
	n.targetList.Delete(connId)
}

func (n *broadcastGroup) ClearAllBroadcastTarget() {
	n.targetList.Range(func(key, value any) bool {
		if conn, err := GetInstanceConnManager().Get(value.(int)); err != nil {
			conn.ExitBroadcastGroup(n.GetGroupId())
		}
		return true
	})
}

func (n *broadcastGroup) BroadcastAllTargets(msgId int32, data proto.Message) {
	n.targetList.Range(func(key, value any) bool {
		if conn, err := GetInstanceConnManager().Get(value.(int)); err != nil {
			conn.SendMsg(msgId, data)
		}
		return true
	})
}
