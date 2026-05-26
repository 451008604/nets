package nets

import (
	"encoding/binary"
	"fmt"
)

type dataPack struct{}

func NewDataPack() IDataPack {
	return &dataPack{}
}

func (d *dataPack) GetHeadLen() int {
	// id int(2 bytes) + dataLen int(2 bytes) / id int(2字节) + dataLen int(2字节)
	return 4
}

func (d *dataPack) Pack(msg IMessage) []byte {
	data := msg.GetData()
	headLen := d.GetHeadLen()
	packed := make([]byte, headLen+len(data))

	// Directly write msgId (2 bytes, little-endian) / 直接写msgId (2字节, 小端)
	binary.LittleEndian.PutUint16(packed[0:2], msg.GetMsgId())
	// Directly write dataLen (2 bytes, little-endian) / 直接写dataLen (2字节, 小端)
	binary.LittleEndian.PutUint16(packed[2:4], msg.GetDataLen())
	// Directly copy data / 直接拷贝data
	copy(packed[headLen:], data)

	return packed
}

func (d *dataPack) UnPack(binaryData []byte) IMessage {
	msgData := GetMessage()

	// Directly read msgId (2 bytes, little-endian) / 直接读msgId (2字节, 小端)
	msgData.Id = binary.LittleEndian.Uint16(binaryData[0:2])
	// Directly read dataLen (2 bytes, little-endian) / 直接读dataLen (2字节, 小端)
	msgData.DataLen = binary.LittleEndian.Uint16(binaryData[2:4])

	// Check if data length exceeds limit / 检查数据长度是否超出限制
	if defaultServer.AppConf.MaxPackSize > 0 && int(msgData.GetDataLen()) > defaultServer.AppConf.MaxPackSize {
		fmt.Printf("received data length exceeds the limit. MaxPackSize %v, msgDataLen %v\n", defaultServer.AppConf.MaxPackSize, msgData.GetDataLen())
		PutMessage(msgData)
		return nil
	}

	totalLen := d.GetHeadLen() + int(msgData.GetDataLen())
	if len(binaryData) >= totalLen {
		msgData.SetData(binaryData[d.GetHeadLen():totalLen])
	}
	return msgData
}
