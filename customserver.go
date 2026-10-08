package nets

import "log/slog"

var defaultServer = &CustomServer{
	AppConf:  GetServerConf(),
	DataPack: NewDataPack(),
	Logger:   newDefaultLogger(slog.LevelInfo),
}

// Custom Server / 自定义服务器
type CustomServer struct {
	AppConf  AppConf   // Service Startup Configuration / 服务启动配置
	DataPack IDataPack // Custom Encoder/Decoder / 自定义编码/解码器
	// Logger is the slog target used by the SDK. A caller-provided logger keeps its handler,
	// level and attributes (including attributes added by Logger.With).
	// Logger 是 SDK 使用的 slog 目标实例；传入后会保留其 Handler、级别及 With 属性。
	Logger *slog.Logger
}

// getLogger returns the configured SDK logger.
// 获取 SDK 当前配置的日志实例。
func getLogger() *slog.Logger {
	defaultServerManagerMu.Lock()
	defer defaultServerManagerMu.Unlock()
	if defaultServer.Logger != nil {
		return defaultServer.Logger
	}
	return slog.Default()
}

// SetCustomServer applies the caller-provided configuration.
// Must precede all global manager/server getters; later configuration is rejected.
// Callers should obtain the defaults via GetServerConf() and mutate the fields they need,
// then pass the result here. This avoids the "zero value means unset" ambiguity, so values
// like ProtocolIsJson=false or Port=0 are honored as-is.
//
// SetCustomServer 应用调用方提供的配置。
// 调用方应通过 GetServerConf() 获取默认值并按需修改字段后传入。
// 这样可避免"零值即未设置"的歧义，使 ProtocolIsJson=false、Port=0 等值被如实采用。
func SetCustomServer(custom *CustomServer) {
	if err := ConfigureDefault(custom); err != nil {
		getLogger().Error("configuration rejected", "err", err)
	}
}

// ConfigureDefault must precede any global manager/server getter. ServerManager configuration
// is immutable after construction; use NewServerManager for another configuration.
func ConfigureDefault(custom *CustomServer) error {
	defaultServerManagerMu.Lock()
	defer defaultServerManagerMu.Unlock()
	if custom == nil {
		return nil
	}
	if defaultServerManager != nil {
		return ErrServerManagerStarted
	}
	validated, err := NewServerManager(custom)
	if err != nil {
		return err
	}
	validated.cancel()
	defaultServer.AppConf = validated.config.AppConf
	if custom.DataPack != nil {
		defaultServer.DataPack = custom.DataPack
	}
	if custom.Logger != nil {
		defaultServer.Logger = custom.Logger
	} else {
		defaultServer.Logger = newDefaultLogger(custom.AppConf.LogLevel)
	}
	return nil
}
