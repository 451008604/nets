package network

import (
	"github.com/451008604/socketServerFrame/iface"
	pb "github.com/451008604/socketServerFrame/proto/bin"
)

type NotifyManager struct {
	notifyList map[string]iface.INotify
}

func NewNotifyManager() iface.INotifyManager {
	return &NotifyManager{
		notifyList: map[string]iface.INotify{},
	}
}

func (n *NotifyManager) GetNotifyGroup() map[string]iface.INotify {
	return n.notifyList
}

func (n *NotifyManager) DelNotifyGroupByID(notifyID string) {
	delete(n.notifyList, notifyID)
}

func (n *NotifyManager) SendNotifyData(notifyID string, msgID pb.MessageID, data []byte) {
	if notify, ok := n.notifyList[notifyID]; ok {
		for _, conn := range notify.GetNotifyTargets() {
			go conn.SendMsg(msgID, data)
		}
	}
}
