package network

import (
	"github.com/451008604/nets/iface"
	"google.golang.org/protobuf/proto"
	"sync"
)

type BroadcastData struct {
	groupID int64
	msgID   int32
	msgData proto.Message
}

func (b *BroadcastData) GroupID() int64 {
	return b.groupID
}

func (b *BroadcastData) SetGroupID(groupID int64) {
	b.groupID = groupID
}

func (b *BroadcastData) MsgID() int32 {
	return b.msgID
}

func (b *BroadcastData) SetMsgID(msgID int32) {
	b.msgID = msgID
}

func (b *BroadcastData) MsgData() proto.Message {
	return b.msgData
}

func (b *BroadcastData) SetMsgData(msgData proto.Message) {
	b.msgData = msgData
}

type BroadcastGroup struct {
	groupID    int64
	targetList sync.Map
}

func (n *BroadcastGroup) GetGroupID() int64 {
	return n.groupID
}

func (n *BroadcastGroup) SetBroadcastTarget(conn iface.IConnection) {
	n.targetList.Store(conn.GetConnID(), conn)
}

func (n *BroadcastGroup) GetBroadcastTarget(connID int) (iface.IConnection, bool) {
	value, ok := n.targetList.Load(connID)
	return value.(iface.IConnection), ok
}

func (n *BroadcastGroup) DelBroadcastTarget(connID int) {
	n.targetList.Delete(connID)
}

func (n *BroadcastGroup) BroadcastAllTargets(connID int, msgID int32, data proto.Message) {
	n.targetList.Range(func(key, value any) bool {
		if connID == key {
			return true
		}
		value.(iface.IConnection).SetNotifyGroupCh(&BroadcastData{groupID: n.GetGroupID(), msgID: msgID, msgData: data})
		return true
	})
}
