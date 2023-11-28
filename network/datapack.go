package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/451008604/nets/config"
	"github.com/451008604/nets/iface"
)

type dataPack struct{}

func NewDataPack() iface.IDataPack {
	return &dataPack{}
}

func (d *dataPack) GetHeadLen() int {
	// totalLen(2字节) + id int(2字节) + dataLen int(2字节)
	return 6
}

func (d *dataPack) Pack(msg iface.IMessage) []byte {
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

func (d *dataPack) Unpack(binaryData []byte) iface.IMessage {
	dataBuff := bytes.NewReader(binaryData)
	msgData := &message{}

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
	if config.GetServerConf().MaxPackSize > 0 && int(msgData.GetDataLen()) > config.GetServerConf().MaxPackSize {
		fmt.Printf("received data length exceeds the limit. MaxPackSize %v, msgDataLen %v\n", config.GetServerConf().MaxPackSize, msgData.GetDataLen())
		return nil
	}
	return msgData
}
