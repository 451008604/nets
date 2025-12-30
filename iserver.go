package main

// 定义服务器接口
type IServer interface {
	// 获取服务器名称
	GetServerName() string
	// 启动服务器
	Start()
}
