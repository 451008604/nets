package nets

import (
	"log/slog"
	"os"
)

// init initializes the global slog logger with a JSON handler writing to stdout
// (AddSource enabled) at Info level as a fallback. The effective level is re-applied
// from AppConf.LogLevel when SetCustomServer is called.
// 初始化全局 slog 日志，默认 JSON 输出到 stdout（含调用源），Info 级别兜底；
// 实际级别在 SetCustomServer 时按 AppConf.LogLevel 重新应用。
func init() {
	applySlogHandler(slog.LevelInfo)
}

// applySlogHandler (re)builds the global slog logger with the given level.
// 依据指定级别（重新）构建全局 slog 日志。
func applySlogHandler(level slog.Level) {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
	})))
}

type AppConf struct {
	AppName          string     // Service Name / 服务名称
	MaxPackSize      uint       // Max Packet Length / 数据包最大长度
	MaxConn          uint       // Max Connections / 最大允许连接数
	WorkerTaskMaxLen uint       // Max Tasks per Worker Queue / 每个工作队列可执行最大任务数量
	WorkerPoolSize   uint       // Worker Pool Size (default CPU*10) / 协程池工作协程数量（默认 CPU*10）
	MaxMsgChanLen    uint       // Max Message Channel Length / 读写通道最大限度
	MaxFlowSecond    int        // Max Requests per Second / 每秒允许的最大请求数量
	ProtocolIsJson   bool       // Use JSON Protocol / 是否使用json协议
	ConnRWTimeOut    uint       // Connection Read/Write Timeout (seconds) / 连接读写超时时间(秒)
	LogLevel         slog.Level // Log Level (default Info) / 日志级别（默认 Info）
	ServerTCP        ServerConf // TCP Service / tcp服务
	ServerWS         ServerConf // WebSocket Service / websocket服务
	ServerHTTP       ServerConf // HTTP Service / http服务
	ServerKCP        ServerConf // KCP Service / kcp服务
}

type ServerConf struct {
	Address     string // IP Address / IP地址
	Port        int    // Port / 端口
	TLSCertPath string // SSL Certificate Path / ssl证书路径
	TLSKeyPath  string // SSL Key Path / ssl密钥路径
}

// Default Configuration / 默认配置
var appConf = &AppConf{
	AppName:          "nets",
	MaxPackSize:      4096,
	MaxConn:          100000,
	WorkerTaskMaxLen: 100,
	MaxMsgChanLen:    100,
	MaxFlowSecond:    -1,
	ProtocolIsJson:   true,
	ConnRWTimeOut:    30,
	ServerTCP: ServerConf{
		Port: 17001,
	},
	ServerWS: ServerConf{
		Port: 17002,
	},
	ServerHTTP: ServerConf{
		Port: 17003,
	},
	ServerKCP: ServerConf{
		Port: 17004,
	},
}

// Get Default Configuration / 获取默认配置
func GetServerConf() AppConf {
	return *appConf
}
