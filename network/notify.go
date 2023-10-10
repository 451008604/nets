package network

import (
	"context"
	"github.com/451008604/socketServerFrame/iface"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
	"sync"
)

type NotifyData struct {
	GroupID int64
	MsgID   pb.MSgID
	MsgData proto.Message
}

type NotifyGroup struct {
	groupID    int64
	targetList sync.Map
	notifyCtx  context.Context
}

func (n *NotifyGroup) GetGroupID() int64 {
	return n.groupID
}

func (n *NotifyGroup) GetNotifyCtx() context.Context {
	return n.notifyCtx
}

func (n *NotifyGroup) SetNotifyTarget(conn iface.IConnection) {
	n.targetList.Store(conn.GetConnID(), conn)

	n.targetList.Range(func(key, value any) bool {
		c := value.(iface.IConnection)
		println(c.RemoteAddrStr())
		return true
	})
}

func (n *NotifyGroup) GetNotifyTarget(connID int) (iface.IConnection, bool) {
	value, ok := n.targetList.Load(connID)
	return value.(iface.IConnection), ok
}

func (n *NotifyGroup) DelNotifyTarget(connID int) {
	n.targetList.Delete(connID)
}

func (n *NotifyGroup) NotifyAllTargets(msgID pb.MSgID, data proto.Message) {
	n.notifyCtx = context.WithValue(context.Background(), "notify", &NotifyData{
		GroupID: n.GetGroupID(),
		MsgID:   msgID,
		MsgData: data,
	})
}
