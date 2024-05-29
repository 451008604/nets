package config

type AppConf struct {
	AppName          string     // 服务名称
	MaxPackSize      int        // 数据包最大长度
	MaxConn          int        // 最大允许连接数
	WorkerPoolSize   int        // 工作队列最大长度
	WorkerTaskMaxLen int        // 每个工作队列可执行最大任务数量
	MaxMsgChanLen    int        // 读写通道最大限度
	ProtocolIsJson   bool       // 是否使用json协议
	ServerTCP        ServerConf // tcp服务
	ServerWS         ServerConf // websocket服务
}

type ServerConf struct {
	Address     string // IP地址
	Port        string // 端口
	TLSCertPath string // ssl证书
	TLSKeyPath  string // ssl密钥
}

var appConf *AppConf

// 默认配置
func init() {
	appConf = &AppConf{
		AppName:          "nets",
		MaxPackSize:      4096,
		MaxConn:          100000,
		WorkerPoolSize:   10000,
		WorkerTaskMaxLen: 100,
		MaxMsgChanLen:    100,
		ProtocolIsJson:   true,
		ServerTCP: ServerConf{
			Port: "17001",
		},
		ServerWS: ServerConf{
			Port: "17002",
		},
	}
}

// 获取默认配置
func GetServerConf() *AppConf {
	return appConf
}
