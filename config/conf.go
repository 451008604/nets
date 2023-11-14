package config

import (
	"fmt"
)

type AppConf struct {
	AppName          string // 服务名称
	MaxPackSize      int    // 数据包最大长度
	MaxConn          int    // 最大允许连接数
	WorkerPoolSize   int    // 工作队列最大长度
	WorkerTaskMaxLen int    // 每个工作队列可执行最大任务数量
	MaxMsgChanLen    int    // 读写通道最大限度
	ProtocolIsJson   bool   // 是否使用json协议
	ServerTCP        ServerConf
	ServerWS         ServerConf
}

type ServerConf struct {
	Address     string // IP地址
	Port        string // 端口
	TLSCertPath string // ssl证书
	TLSKeyPath  string // ssl密钥
}

var appConf AppConf

// 初始化服务器配置
func SetServerConf(conf AppConf) {
	appConf = conf
}

// 获取全局配置对象
func GetServerConf() AppConf {
	if appConf.AppName == "" {
		fmt.Printf("server config not init\n")
	}
	return appConf
}
