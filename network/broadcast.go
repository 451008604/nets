package network

import (
	"github.com/451008604/nets/iface"
	"google.golang.org/protobuf/proto"
	"sync"
)

type broadcastData struct {
	groupID int64
	msgID   int32
	msgData proto.Message
}

func (b *broadcastData) GroupID() int64 {
	return b.groupID
}

func (b *broadcastData) SetGroupID(groupID int64) {
	b.groupID = groupID
}

func (b *broadcastData) MsgID() int32 {
	return b.msgID
}

func (b *broadcastData) SetMsgID(msgID int32) {
	b.msgID = msgID
}

func (b *broadcastData) MsgData() proto.Message {
	return b.msgData
}

func (b *broadcastData) SetMsgData(msgData proto.Message) {
	b.msgData = msgData
}

type broadcastGroup struct {
	groupID    int64
	targetList sync.Map
}

func (n *broadcastGroup) GetGroupID() int64 {
	return n.groupID
}

func (n *broadcastGroup) SetBroadcastTarget(conn iface.IConnection) {
	n.targetList.Store(conn.GetConnID(), conn)
}

func (n *broadcastGroup) GetBroadcastTarget(connID int) (iface.IConnection, bool) {
	value, ok := n.targetList.Load(connID)
	return value.(iface.IConnection), ok
}

func (n *broadcastGroup) DelBroadcastTarget(connID int) {
	n.targetList.Delete(connID)
}

func (n *broadcastGroup) BroadcastAllTargets(connID int, msgID int32, data proto.Message) {
	n.targetList.Range(func(key, value any) bool {
		if connID == key {
			return true
		}
		value.(iface.IConnection).SetNotifyGroupCh(&broadcastData{groupID: n.GetGroupID(), msgID: msgID, msgData: data})
		return true
	})
}
