package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

type dataPack struct{}

func NewDataPack() IDataPack {
	return &dataPack{}
}

func (d *dataPack) GetHeadLen() int {
	// id int(2字节) + dataLen int(2字节)
	return 4
}

func (d *dataPack) Pack(msg IMessage) []byte {
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

func (d *dataPack) UnPack(binaryData []byte) IMessage {
	dataBuff := bytes.NewReader(binaryData)
	msgData := &Message{}
	// 读msgId
	if binary.Read(dataBuff, binary.LittleEndian, &msgData.Id) != nil {
		return nil
	}
	// 读dataLen
	if binary.Read(dataBuff, binary.LittleEndian, &msgData.DataLen) != nil {
		return nil
	}
	// 检查数据长度是否超出限制
	if defaultServer.AppConf.MaxPackSize > 0 && int(msgData.GetDataLen()) > defaultServer.AppConf.MaxPackSize {
		fmt.Printf("received data length exceeds the limit. MaxPackSize %v, msgDataLen %v\n", defaultServer.AppConf.MaxPackSize, msgData.GetDataLen())
		return nil
	}

	totalLen := d.GetHeadLen() + int(msgData.GetDataLen())
	if len(binaryData) >= totalLen {
		msgData.SetData(binaryData[d.GetHeadLen():totalLen])
	}
	return msgData
}
