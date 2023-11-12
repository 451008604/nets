package config

import (
	"encoding/json"
	"os"
)

type GlobalConf struct {
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

var conf GlobalConf

func initServerConfig() {
	readFile, err := os.ReadFile("conf.json")
	if err != nil {
		return
	}

	err = json.Unmarshal(readFile, &conf)
	if err != nil {
		return
	}
}

// GetGlobalObject 获取全局配置对象
func GetGlobalObject() GlobalConf {
	if conf.AppName == "" {
		initServerConfig()
	}
	return conf
}
