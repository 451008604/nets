package nets

var defaultServer = &CustomServer{
	AppConf:  GetServerConf(),
	DataPack: NewDataPack(),
}

// Custom Server / 自定义服务器
type CustomServer struct {
	AppConf  *AppConf  // Service Startup Configuration / 服务启动配置
	DataPack IDataPack // Custom Encoder/Decoder / 自定义编码/解码器
}

// SetCustomServer applies the caller-provided configuration.
// Callers should obtain the defaults via GetServerConf() and mutate the fields they need,
// then pass the result here. This avoids the "zero value means unset" ambiguity, so values
// like ProtocolIsJson=false or Port=0 are honored as-is.
//
// SetCustomServer 应用调用方提供的配置。
// 调用方应通过 GetServerConf() 获取默认值并按需修改字段后传入。
// 这样可避免“零值即未设置”的歧义，使 ProtocolIsJson=false、Port=0 等值被如实采用。
func SetCustomServer(custom *CustomServer) {
	if custom == nil {
		return
	}
	if custom.AppConf != nil {
		defaultServer.AppConf = custom.AppConf
	}
	if custom.DataPack != nil {
		defaultServer.DataPack = custom.DataPack
	}
}
