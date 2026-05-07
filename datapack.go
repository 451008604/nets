package nets

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"sync"
)

var bufferPool = sync.Pool{
	New: func() any { return bytes.NewBuffer(make([]byte, 0, 64)) },
}

type dataPack struct{}

func NewDataPack() IDataPack {
	return &dataPack{}
}

func (d *dataPack) GetHeadLen() int {
	// id int(2字节) + dataLen int(2字节)
	return 4
}

func (d *dataPack) Pack(msg IMessage) []byte {
	dataBuff := bufferPool.Get().(*bytes.Buffer)
	dataBuff.Reset()
	defer bufferPool.Put(dataBuff)

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
	// 拷贝结果，防止 buffer 归还后数据被覆盖
	return append([]byte(nil), dataBuff.Bytes()...)
}

func (d *dataPack) UnPack(binaryData []byte) IMessage {
	dataBuff := bytes.NewReader(binaryData)
	msgData := GetMessage()
	// 读msgId
	if binary.Read(dataBuff, binary.LittleEndian, &msgData.Id) != nil {
		PutMessage(msgData)
		return nil
	}
	// 读dataLen
	if binary.Read(dataBuff, binary.LittleEndian, &msgData.DataLen) != nil {
		PutMessage(msgData)
		return nil
	}
	// 检查数据长度是否超出限制
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
