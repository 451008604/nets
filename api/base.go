package api

import "github.com/451008604/socketServerFrame/network"

func init() {
	// 注册路由
	RegisterRouter(network.GetInstanceMsgHandler())
}
