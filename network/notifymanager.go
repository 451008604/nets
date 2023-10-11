package network

import (
	"github.com/451008604/socketServerFrame/iface"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"sync"
)

type BroadcastManager struct {
	notifyList      sync.Map
	globalBroadcast iface.IBroadcast
}

var insBroadcastManager *BroadcastManager
var insBroadcastManagerOnce = sync.Once{}

func GetInsBroadcastManager() iface.IBroadcastManager {
	insBroadcastManagerOnce.Do(func() {
		insBroadcastManager = &BroadcastManager{
			notifyList: sync.Map{},
		}
		insBroadcastManager.globalBroadcast = insBroadcastManager.NewBroadcastGroup()
	})
	return insBroadcastManager
}

// 新建广播组
func (n *BroadcastManager) NewBroadcastGroup() iface.IBroadcast {
	broadcast := &BroadcastGroup{
		groupID:    int64(10000000000) + int64(uuid.New().ID()),
		targetList: sync.Map{},
	}
	n.notifyList.Store(broadcast.GetGroupID(), broadcast)
	return broadcast
}

func (n *BroadcastManager) GetGlobalBroadcast() iface.IBroadcast {
	return n.globalBroadcast
}

func (n *BroadcastManager) GetBroadcast(groupID int64) (any, bool) {
	value, ok := n.notifyList.Load(groupID)
	return value.(iface.IBroadcast), ok
}

func (n *BroadcastManager) DelBroadcastByID(groupID int64) {
	n.notifyList.Delete(groupID)
}

func (n *BroadcastManager) SendBroadcastData(groupID int64, msgID pb.MSgID, data proto.Message) {
	if broadcast, ok := n.GetBroadcast(groupID); ok {
		broadcast.(iface.IBroadcast).BroadcastAllTargets(msgID, data)
	}
}
