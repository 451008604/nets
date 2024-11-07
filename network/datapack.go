package network

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/451008604/nets/iface"
)

type dataPack struct{}

func NewDataPack() iface.IDataPack {
	return &dataPack{}
}

func (d *dataPack) getHeadLen() int {
	// id int(2字节) + dataLen int(2字节)
	return 4
}

func (d *dataPack) Pack(msg iface.IMessage) []byte {
	dataBuff := bytes.NewBuffer([]byte{})

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

func (d *dataPack) UnPack(binaryData []byte) iface.IMessage {
	dataBuff := bytes.NewReader(binaryData)
	msgData := &message{}
	// 读msgId
	if binary.Read(dataBuff, binary.LittleEndian, &msgData.id) != nil {
		return nil
	}
	// 读dataLen
	if binary.Read(dataBuff, binary.LittleEndian, &msgData.dataLen) != nil {
		return nil
	}
	// 检查数据长度是否超出限制
	if defaultServer.AppConf.MaxPackSize > 0 && int(msgData.GetDataLen()) > defaultServer.AppConf.MaxPackSize {
		fmt.Printf("received data length exceeds the limit. MaxPackSize %v, msgDataLen %v\n", defaultServer.AppConf.MaxPackSize, msgData.GetDataLen())
		return nil
	}

	totalLen := d.getHeadLen() + int(msgData.GetDataLen())
	if len(binaryData) >= totalLen {
		msgData.SetData(binaryData[d.getHeadLen():totalLen])
	}
	return msgData
}
