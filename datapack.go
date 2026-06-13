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

	// Reject oversize payloads instead of silently truncating DataLen to 65535.
	// Truncation would cause the receiver to misread trailing bytes as the next
	// message, corrupting the entire subsequent stream.
	if len(data) > 0xFFFF {
		fmt.Printf("pack aborted: data length %d exceeds uint16 max (%d)\n", len(data), 0xFFFF)
		return nil
	}

	packed := make([]byte, headLen+len(data))

	// Directly write msgId (2 bytes, little-endian) / 直接写msgId (2字节, 小端)
	binary.LittleEndian.PutUint16(packed[0:2], msg.GetMsgId())
	// Directly write dataLen (2 bytes, little-endian) / 直接写dataLen (2字节, 小端)
	binary.LittleEndian.PutUint16(packed[2:4], uint16(len(data)))
	// Directly copy data / 直接拷贝数据
	copy(packed[headLen:], data)

	return packed
}

func (d *dataPack) UnPack(binaryData []byte) IMessage {
	if len(binaryData) < d.GetHeadLen() {
		return nil
	}

	msgData := GetMessage()

	// Directly read msgId (2 bytes, little-endian) / 直接读msgId (2字节, 小端)
	msgData.Id = binary.LittleEndian.Uint16(binaryData[0:2])
	// Directly read dataLen (2 bytes, little-endian) / 直接读dataLen (2字节, 小端)
	msgData.DataLen = binary.LittleEndian.Uint16(binaryData[2:4])

	// Check if data length exceeds limit / 检查数据长度是否超出限制
	maxPackSize := defaultServer.AppConf.MaxPackSize
	if maxPackSize <= 0 {
		maxPackSize = 4096 // Default limit / 默认限制
	}
	if int(msgData.GetDataLen()) > maxPackSize {
		fmt.Printf("received data length exceeds the limit. MaxPackSize %v, msgDataLen %v\n", maxPackSize, msgData.GetDataLen())
		PutMessage(msgData)
		return nil
	}

	totalLen := d.GetHeadLen() + int(msgData.GetDataLen())
	if len(binaryData) >= totalLen {
		msgData.SetData(binaryData[d.GetHeadLen():totalLen])
	}
	return msgData
}
