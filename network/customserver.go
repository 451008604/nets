package network

import (
	"github.com/451008604/nets/config"
	"github.com/451008604/nets/iface"
)

var defaultServer *CustomServer

// 自定义服务器
type CustomServer struct {
	AppConf    *config.AppConf // 服务启动配置
	DataPacket iface.IDataPack // 编码/解码器
}

func init() {
	defaultServer = &CustomServer{
		AppConf:    config.GetServerConf(),
		DataPacket: NewDataPack(),
	}
}

func setCustomServer(custom *CustomServer) {
	if custom.AppConf != nil {
		defaultServer.AppConf = custom.AppConf
	}
	if custom.DataPacket != nil {
		defaultServer.DataPacket = custom.DataPacket
	}
}
