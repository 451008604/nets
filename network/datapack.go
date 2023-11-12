package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/451008604/nets/config"
	"github.com/451008604/nets/iface"
)

type DataPack struct{}

// 新数据包
func NewDataPack() *DataPack {
	return &DataPack{}
}

func (d *DataPack) GetHeadLen() int {
	// totalLen(2字节) + id int(2字节) + dataLen int(2字节)
	return 6
}

// 封包
func (d *DataPack) Pack(msg iface.IMessage) []byte {
	dataBuff := bytes.NewBuffer([]byte{})

	// 写totalLen
	if binary.Write(dataBuff, binary.LittleEndian, msg.GetTotalLen()) != nil {
		return nil
	}
	// 写msgId
	if binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgId()) != nil {
		return nil
	}
	// 写dataLen
	if binary.Write(dataBuff, binary.LittleEndian, msg.GetDataLen()) != nil {
		return nil
	}
	// 写data数据
	if binary.Write(dataBuff, binary.LittleEndian, msg.GetData()) != nil {
		return nil
	}
	return dataBuff.Bytes()
}

// 拆包(只获取到包头Id,dataLen)
func (d *DataPack) Unpack(binaryData []byte) iface.IMessage {
	dataBuff := bytes.NewReader(binaryData)
	msgData := &Message{}

	// 读totalLen
	if binary.Read(dataBuff, binary.LittleEndian, &msgData.totalLen) != nil {
		return nil
	}
	// 读msgId
	if binary.Read(dataBuff, binary.LittleEndian, &msgData.id) != nil {
		return nil
	}
	// 读dataLen
	if binary.Read(dataBuff, binary.LittleEndian, &msgData.dataLen) != nil {
		return nil
	}
	// 检查数据长度是否超出限制
	if config.GetGlobalObject().MaxPackSize > 0 && int(msgData.GetDataLen()) > config.GetGlobalObject().MaxPackSize {
		fmt.Printf("received data length exceeds the limit. MaxPackSize %v, msgDataLen %v\n", config.GetGlobalObject().MaxPackSize, msgData.GetDataLen())
		return nil
	}
	return msgData
}
