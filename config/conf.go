package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/451008604/socketServerFrame/logs"
)

var configPath string // 配置的文件夹路径

type GlobalObj struct {
	Debug            bool   // 是否Debug模式
	Name             string // 当前服务名称
	Version          string // 当前服务版本号
	MaxPackSize      int    // 传输数据包最大值
	MaxConn          int    // 当前服务允许的最大连接数
	WorkerPoolSize   int    // work池大小
	WorkerTaskMaxLen int    // work对应的执行队列内任务数量的上限
	MaxMsgChanLen    int    // 读写消息的通道最大缓冲数
	ProtocolIsJson   bool   // 是否采用Json协议
	HostTCP          string // TCP服务地址
	PortTCP          string // TCP服务端口
	HostWS           string // WS服务地址
	PortWS           string // WS服务端口
	TLSCertPath      string // TLS证书路径
	TLSKeyPath       string // TLS密钥路径
	RedisAddress     string // Redis地址
	RedisPassWord    string // Redis密码
}

var globalObject *GlobalObj

func init() {
	globalObject = &GlobalObj{
		Debug:            false,
		Name:             "MyProject",
		Version:          "v0.1",
		MaxPackSize:      4096,
		MaxConn:          10,
		WorkerPoolSize:   1000,
		WorkerTaskMaxLen: 1000,
		MaxMsgChanLen:    100,
		ProtocolIsJson:   true,
	}

	globalObject.Reload()
	logs.SetPrintMode(globalObject.Debug)

	str, _ := json.Marshal(globalObject)
	logs.PrintLogInfo(fmt.Sprintf("服务配置参数：%v", string(str)))
}

// GetGlobalObject 获取全局配置对象
func GetGlobalObject() GlobalObj {
	return *globalObject
}

func (o *GlobalObj) Reload() {
	err := json.Unmarshal(getConfigDataToBytes("./config/", "config.json"), &globalObject)
	logs.PrintLogErr(err)
}

// 获取配置数据到字节
func getConfigDataToBytes(configPath string, configName string) []byte {
	bytes, err := ioutil.ReadFile(configPath + configName)
	logs.PrintLogPanic(err)
	return bytes
}
