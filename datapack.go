package nets

import (
	"encoding/binary"
	"fmt"
)

type dataPack struct{ maxPackSize uint }

func NewDataPack() IDataPack        { return &dataPack{maxPackSize: GetServerConf().MaxPackSize} }
func (d *dataPack) GetHeadLen() int { return 4 }
func (d *dataPack) PackWithError(msg IMessage) ([]byte, error) {
	if msg == nil {
		return nil, ErrInvalidMessage
	}
	data := msg.GetData()
	if len(data) > 65535 || uint(len(data)) > d.maxPackSize {
		return nil, fmt.Errorf("%w: payload too large", ErrInvalidMessage)
	}
	packed := make([]byte, d.GetHeadLen()+len(data))
	binary.LittleEndian.PutUint16(packed[0:2], msg.GetMsgId())
	binary.LittleEndian.PutUint16(packed[2:4], uint16(len(data)))
	copy(packed[d.GetHeadLen():], data)
	return packed, nil
}
func (d *dataPack) Pack(msg IMessage) []byte {
	packed, err := d.PackWithError(msg)
	if err != nil {
		getLogger().Warn("pack failed", "err", err)
	}
	return packed
}
func (d *dataPack) UnPackWithError(data []byte) (IMessage, error) {
	if len(data) < d.GetHeadLen() {
		return nil, fmt.Errorf("%w: incomplete header", ErrInvalidMessage)
	}
	size := binary.LittleEndian.Uint16(data[2:4])
	if uint(size) > d.maxPackSize {
		return nil, fmt.Errorf("%w: payload too large", ErrInvalidMessage)
	}
	msg := GetMessage()
	msg.Id = binary.LittleEndian.Uint16(data[0:2])
	msg.DataLen = size
	total := d.GetHeadLen() + int(size)
	if len(data) >= total {
		msg.SetData(data[d.GetHeadLen():total])
	}
	return msg, nil
}
func (d *dataPack) UnPack(data []byte) IMessage {
	msg, _ := d.UnPackWithError(data)
	return msg
}
