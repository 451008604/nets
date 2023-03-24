package network

import "github.com/451008604/socketServerFrame/iface"

type Notify struct {
	NotifyID   string
	targetList []iface.IConnection
}

func (n *Notify) SetNotifyTarget(conn iface.IConnection) {
	if n.HasNotifyTarget(conn) {
		return
	}
	n.targetList = append(n.targetList, conn)
}

func (n *Notify) GetNotifyTargets() (connList []iface.IConnection) {
	return n.targetList
}

func (n *Notify) DelNotifyTarget(conn iface.IConnection) {
	for i := range n.targetList {
		if n.targetList[i] == conn {
			n.targetList = append(n.targetList[:i], n.targetList[i+1:]...)
			return
		}
	}
}

func (n *Notify) HasNotifyTarget(conn iface.IConnection) bool {
	for i := range n.targetList {
		if n.targetList[i] == conn {
			return true
		}
	}
	return false
}
