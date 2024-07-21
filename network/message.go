package network

import (
	"fmt"
	"github.com/451008604/nets/iface"
)

type message struct {
	id      uint16 // 消息Id
	dataLen uint16 // 消息长度
	data    []byte // 消息内容
}

func NewMsgPackage(id int32, data []byte) iface.IMessage {
	return &message{
		id:      uint16(id),
		dataLen: uint16(len(data)),
		data:    data,
	}
}

func (m *message) GetDataLen() uint16 {
	return m.dataLen
}

func (m *message) SetDataLen(u uint16) {
	m.dataLen = u
}

func (m *message) GetMsgId() uint16 {
	return m.id
}

func (m *message) SetMsgId(u uint16) {
	m.id = u
}

func (m *message) GetData() []byte {
	return m.data
}

func (m *message) SetData(bytes []byte) {
	m.data = bytes
}

func (m *message) TaskHandler(conn iface.IConnection) {
	iMsgHandler := GetInstanceMsgHandler()
	defer iMsgHandler.GetErrCapture(conn)

	// 连接关闭时丢弃后续所有操作
	if conn.IsClose() {
		return
	}

	router, ok := iMsgHandler.GetApis()[int32(m.GetMsgId())]
	if !ok {
		return
	}

	msgData := router.GetNewMsg()
	if err := conn.ByteToProtocol(m.GetData(), msgData); err != nil {
		fmt.Printf("api msgId %v parsing %v error %v\n", m.GetMsgId(), m.GetData(), err)
		return
	}

	// 限流控制
	if conn.FlowControl() {
		fmt.Printf("flowControl RemoteAddress: %v, GetMsgId: %v, GetData: %v\n", conn.RemoteAddrStr(), m.GetMsgId(), m.GetData())
		return
	}

	// 过滤器校验
	if iMsgHandler.GetFilter() != nil && !iMsgHandler.GetFilter()(conn, msgData) {
		return
	}

	// 对应的逻辑处理方法
	router.RunHandler(conn, msgData)
}
