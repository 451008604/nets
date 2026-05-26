package nets

// Pack/Unpack, obtain message data through fixed packet header to solve TCP sticky packet problem / 封包拆包，通过固定的包头获取消息数据，解决TCP粘包问题
type IDataPack interface {
	// Get Message Header Length / 获取消息头长度
	GetHeadLen() int
	// Message Pack / 消息封包
	Pack(msg IMessage) []byte
	// Message Unpack / 消息拆包
	UnPack([]byte) IMessage
}
