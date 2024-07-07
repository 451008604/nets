package iface

type IServerManager interface {
	// 注册服务
	RegisterServer(server ...IServer)
	// 获取启动的服务列表
	Servers() []IServer
	// 获取服务是否已关闭
	IsClose() bool
	// 增加连接断开等待执行计数
	WaitGroupAdd(delta int)
	// 连接断开hook已完成
	WaitGroupDone()
	// 停止所有服务
	StopAll()
}
