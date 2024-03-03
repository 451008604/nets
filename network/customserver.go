package network

import (
	"github.com/451008604/nets/config"
	"github.com/451008604/nets/iface"
	"reflect"
)

var defaultServer *CustomServer

// 自定义服务器
type CustomServer struct {
	AppConf    *config.AppConf // 服务启动配置
	DataPacket iface.IDataPack // 编码/解码器
}

// 初始化各组件
func init() {
	defaultServer = &CustomServer{
		AppConf:    config.GetServerConf(),
		DataPacket: NewDataPack(),
	}
}

// 设置自定义服务器参数
func SetCustomServer(custom *CustomServer) {
	defaultServer.AppConf = MergeStructValues(defaultServer.AppConf, custom.AppConf)

	if custom.DataPacket != nil {
		defaultServer.DataPacket = custom.DataPacket
	}
}

// 将 defaultData 与 customData 进行合并，相同字段赋值优先使用 customData
func MergeStructValues[T any](defaultData, customData *T) *T {
	if customData == nil {
		if defaultData != nil {
			return defaultData
		}
		return new(T)
	} else if defaultData == nil {
		if customData != nil {
			return customData
		}
		return new(T)
	}
	resultData := new(T)

	v1 := reflect.ValueOf(defaultData).Elem()
	v2 := reflect.ValueOf(customData).Elem()
	v3 := reflect.ValueOf(resultData).Elem()

	for i := 0; i < v1.NumField(); i++ {
		fieldValue1 := v1.Field(i)
		fieldValue2 := v2.Field(i)

		// 如果 customData 中的字段有值则使用它；否则使用 defaultData 中的对应字段值
		if !fieldValue2.IsZero() {
			v3.Field(i).Set(fieldValue2)
		} else {
			v3.Field(i).Set(fieldValue1)
		}
	}
	return resultData
}
