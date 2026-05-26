package nets

// Define Server Interface / 定义服务器接口
type IServer interface {
	// Get Server Name / 获取服务器名称
	GetServerName() string
	// Start Server / 启动服务器
	Start()
}
