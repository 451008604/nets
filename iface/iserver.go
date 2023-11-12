package iface

// 定义服务器接口
type IServer interface {
	// 获取服务器名称
	GetServerName() string
	// 启动服务器
	Start()
	// 停止服务器
	Stop()
	// 开启业务服务
	Listen() bool
	// 获取封包/拆包工具
	DataPacket() IDataPack
}
