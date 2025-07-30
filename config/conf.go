package config

type AppConf struct {
	AppName          string     // 服务名称
	MaxPackSize      int        // 数据包最大长度
	MaxConn          int        // 最大允许连接数
	WorkerPoolSize   int        // 工作池容量
	WorkerTaskMaxLen int        // 每个工作队列可执行最大任务数量
	MaxMsgChanLen    int        // 读写通道最大限度
	MaxFlowSecond    int        // 每秒允许的最大请求数量
	ProtocolIsJson   bool       // 是否使用json协议
	ConnRWTimeOut    int        // 连接读写超时时间(秒)
	ServerTCP        ServerConf // tcp服务
	ServerWS         ServerConf // websocket服务
	ServerHTTP       ServerConf // http服务
	ServerKCP        ServerConf // http服务
}

type ServerConf struct {
	Address     string // IP地址
	Port        string // 端口
	TLSCertPath string // ssl证书路径
	TLSKeyPath  string // ssl密钥路径
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
		MaxFlowSecond:    30,
		ProtocolIsJson:   true,
		ConnRWTimeOut:    300,
		ServerTCP: ServerConf{
			Port: "17001",
		},
		ServerWS: ServerConf{
			Port: "17002",
		},
		ServerHTTP: ServerConf{
			Port: "17003",
		},
		ServerKCP: ServerConf{
			Port: "17004",
		},
	}
}

// 获取默认配置
func GetServerConf() *AppConf {
	return appConf
}
