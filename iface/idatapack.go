package iface

// 封包拆包，通过固定的包头获取消息数据，解决TCP粘包问题
type IDataPack interface {
	// 获取包头长度
	GetHeadLen() int
	// 消息封包
	Pack(msg IMessage) []byte
	// 消息拆包
	UnPack([]byte) IMessage
}
