package iface

// 封包拆包，通过固定的包头获取消息数据，解决TCP粘包问题
type IDataPack interface {
	// 获取包头长度
	GetHeadLen() uint32
	// 消息拆包
	Pack(msg IMessage) []byte
	// 消息封包
	Unpack([]byte) IMessage
}
