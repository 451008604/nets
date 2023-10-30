package network

import (
	"github.com/451008604/socketServerFrame/iface"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
	"sync"
)

type BroadcastData struct {
	groupID int64
	msgID   pb.MSgID
	msgData proto.Message
}

func (b *BroadcastData) GroupID() int64 {
	return b.groupID
}

func (b *BroadcastData) SetGroupID(groupID int64) {
	b.groupID = groupID
}

func (b *BroadcastData) MsgID() pb.MSgID {
	return b.msgID
}

func (b *BroadcastData) SetMsgID(msgID pb.MSgID) {
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

func (n *BroadcastGroup) BroadcastAllTargets(connID int, msgID pb.MSgID, data proto.Message) {
	n.targetList.Range(func(key, value any) bool {
		if connID == key {
			return true
		}
		value.(iface.IConnection).SetNotifyGroupCh(&BroadcastData{groupID: n.GetGroupID(), msgID: msgID, msgData: data})
		return true
	})
}
