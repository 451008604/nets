package network

import (
	"github.com/451008604/socketServerFrame/iface"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
	"sync"
)

type NotifyManager struct {
	notifyList sync.Map //  map[string]iface.INotify
}

func NewNotifyManager() iface.INotifyManager {
	return &NotifyManager{
		notifyList: sync.Map{},
	}
}

func (n *NotifyManager) SetNotify(notify iface.INotify) {
	n.notifyList.Store(notify.GetNotifyID(), notify)
}

func (n *NotifyManager) GetNotify(notifyID uint32) (any, bool) {
	value, ok := n.notifyList.Load(notifyID)
	return value.(iface.INotify), ok
}

func (n *NotifyManager) DelNotifyByID(notifyID uint32) {
	n.notifyList.Delete(notifyID)
}

func (n *NotifyManager) SendNotifyData(notifyID uint32, msgID pb.MSgID, data proto.Message) {
	if notify, ok := n.GetNotify(notifyID); ok {
		notify.(iface.INotify).NotifyAllTargets(msgID, data)
	}
}
