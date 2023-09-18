package config

import (
	"bytes"
	"fmt"
	"github.com/451008604/socketServerFrame/logs"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"io/ioutil"
	"time"
)

var jsonsPath = "./config/jsons/"

type GlobalConf struct {
	Debug            bool
	AppName          string
	Version          string
	MaxPackSize      int
	MaxConn          int
	WorkerPoolSize   int
	WorkerTaskMaxLen int
	MaxMsgChanLen    int
	ProtocolIsJson   bool
	ServerTCP        Server
	ServerWS         Server
	Redis            Database
	Mysql            Database
}

type Server struct {
	Address     string
	Port        string
	TLSCertPath string
	TLSKeyPath  string
}

type Database struct {
	Address  string
	Username string
	Password string
}

var conf GlobalConf

func init() {
	viper.SetConfigType("toml")
	// 注册需要监控的配置文件
	viper.SetConfigFile("./config.toml")
	viper.WatchConfig()
	// 开启监控回调，限制每秒最多执行1次
	t := time.Now().Unix()
	viper.OnConfigChange(func(in fsnotify.Event) {
		now := time.Now().Unix()
		if in.Has(fsnotify.Write) && now-1 > t {
			t = now
			loadServerConfig()
		}
	})

	// 初始化配置内容
	configByte, err := ioutil.ReadFile("./config.toml")
	logs.PrintLogPanic(err)
	logs.PrintLogPanic(viper.ReadConfig(bytes.NewBuffer(configByte)))
	loadServerConfig()
}

// 解析配置内容到结构体
func loadServerConfig() {
	logs.PrintLogErr(viper.Unmarshal(&conf))
	logs.SetPrintMode(conf.Debug)
	logs.PrintLogInfo(fmt.Sprintf("服务配置参数：%v", viper.AllSettings()))
}

// GetGlobalObject 获取全局配置对象
func GetGlobalObject() GlobalConf {
	return conf
}
