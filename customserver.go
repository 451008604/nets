package nets

import (
	"reflect"
)

var defaultServer = &CustomServer{
	AppConf:  GetServerConf(),
	DataPack: NewDataPack(),
	Message:  NewMsgPackage,
}

// 自定义服务器
type CustomServer struct {
	AppConf  *AppConf                             // 服务启动配置
	DataPack IDataPack                            // 自定义编码/解码器
	Message  func(id int32, data []byte) IMessage // 自定消息

}

// 设置自定义服务器参数
func SetCustomServer(custom *CustomServer) {
	defaultServer.AppConf = mergeStructValues(defaultServer.AppConf, custom.AppConf)

	if custom.DataPack != nil {
		defaultServer.DataPack = custom.DataPack
	}

	if custom.Message != nil {
		defaultServer.Message = custom.Message
	}
}

// 将 defaultData 与 customData 进行合并，相同字段赋值优先使用 customData
func mergeStructValues[T any](defaultData, customData *T) *T {
	if customData == nil && defaultData == nil {
		return new(T)
	}
	if customData == nil {
		return defaultData
	} else if defaultData == nil {
		return customData
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
