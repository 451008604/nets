package network

import (
	"github.com/451008604/socketServerFrame/iface"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"sync"
)

type Notify struct {
	NotifyID   uint32
	targetList sync.Map
}

func NewNotify() iface.INotify {
	return &Notify{
		NotifyID:   uuid.New().ID(),
		targetList: sync.Map{},
	}
}

func (n *Notify) GetNotifyID() uint32 {
	return n.NotifyID
}

func (n *Notify) SetNotifyTarget(conn iface.IConnection) {
	n.targetList.Store(conn.GetConnID(), conn)
}

func (n *Notify) GetNotifyTarget(connID uint32) (iface.IConnection, bool) {
	value, ok := n.targetList.Load(connID)
	return value.(iface.IConnection), ok
}

func (n *Notify) DelNotifyTarget(connID uint32) {
	n.targetList.Delete(connID)
}

func (n *Notify) NotifyAllTargets(msgID pb.MSgID, data proto.Message) {
	n.targetList.Range(func(key, value any) bool {
		value.(iface.IConnection).SendMsg(msgID, data)
		return true
	})
}
