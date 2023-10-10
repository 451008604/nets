package iface

import (
	"context"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
)

// 通知管理器
type INotifyManager interface {
	// 新建通知组
	NewNotifyGroup() INotify
	// 获取全局通知
	GetGlobalNotify() INotify
	// 获取通知组
	GetNotify(groupID int64) (any, bool)
	// 根据组ID删除指定通知
	DelNotifyByID(groupID int64)
	// 向通知组内所有对象广播信息
	SendNotifyData(groupID int64, msgID pb.MSgID, data proto.Message)
}

// 通知组
type INotify interface {
	// 获取通知对象
	GetGroupID() int64
	// 获取通知上下文
	GetNotifyCtx() context.Context
	// 设置通知对象
	SetNotifyTarget(conn IConnection)
	// 获取通知对象
	GetNotifyTarget(connID int) (IConnection, bool)
	// 删除一个通知对象
	DelNotifyTarget(connID int)
	// 通知所有对象
	NotifyAllTargets(msgID pb.MSgID, data proto.Message)
}
