package network

import (
	"context"
	"github.com/451008604/socketServerFrame/iface"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"sync"
)

type NotifyManager struct {
	notifyList   sync.Map
	globalNotify iface.INotify // 全局通知组
}

var instanceNotifyManager *NotifyManager
var instanceNotifyManagerOnce = sync.Once{}

func GetInstanceNotifyManager() iface.INotifyManager {
	instanceNotifyManagerOnce.Do(func() {
		instanceNotifyManager = &NotifyManager{
			notifyList: sync.Map{},
		}
		instanceNotifyManager.globalNotify = instanceNotifyManager.NewNotifyGroup()
	})
	return instanceNotifyManager
}

// 新建通知组
func (n *NotifyManager) NewNotifyGroup() iface.INotify {
	notify := &NotifyGroup{
		groupID:    int64(10000000000) + int64(uuid.New().ID()),
		targetList: sync.Map{},
		notifyCtx:  context.Background(),
	}
	n.notifyList.Store(notify.GetGroupID(), notify)
	return notify
}

func (n *NotifyManager) GetGlobalNotify() iface.INotify {
	return n.globalNotify
}

func (n *NotifyManager) GetNotify(groupID int64) (any, bool) {
	value, ok := n.notifyList.Load(groupID)
	return value.(iface.INotify), ok
}

func (n *NotifyManager) DelNotifyByID(groupID int64) {
	n.notifyList.Delete(groupID)
}

func (n *NotifyManager) SendNotifyData(groupID int64, msgID pb.MSgID, data proto.Message) {
	if notify, ok := n.GetNotify(groupID); ok {
		notify.(iface.INotify).NotifyAllTargets(msgID, data)
	}
}
