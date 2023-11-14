package config

import (
	"fmt"
)

type ServerConf struct {
	AppName          string // 服务名称
	MaxPackSize      int    // 数据包最大长度
	MaxConn          int    // 最大允许连接数
	WorkerPoolSize   int    // 工作队列最大长度
	WorkerTaskMaxLen int    // 每个工作队列可执行最大任务数量
	MaxMsgChanLen    int    // 读写通道最大限度
	ProtocolIsJson   bool   // 是否使用json协议
	ServerTCP        Server
	ServerWS         Server
}

type Server struct {
	Address     string // IP地址
	Port        string // 端口
	TLSCertPath string // ssl证书
	TLSKeyPath  string // ssl密钥
}

var serverConf ServerConf

// 初始化服务器配置
func SetServerConf(conf ServerConf) {
	serverConf = conf
}

// 获取全局配置对象
func GetServerConf() ServerConf {
	if serverConf.AppName == "" {
		fmt.Printf("server config not init\n")
	}
	return serverConf
}
