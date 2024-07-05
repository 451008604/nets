package iface

type IServerManager interface {
	// 注册服务
	RegisterServer(server ...IServer)
}
