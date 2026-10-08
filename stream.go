package nets

import (
	"io"
	"net"
)

func readStreamPacket(c IConnection, conn net.Conn) bool {
	r := managerForConnection(c)
	head := make([]byte, r.config.DataPack.GetHeadLen())
	if _, err := io.ReadFull(conn, head); err != nil {
		return false
	}
	msg := r.config.DataPack.UnPack(head)
	if msg == nil {
		return false
	}
	handedOff := false
	defer func() {
		if !handedOff {
			PutMessage(msg)
		}
	}()
	size := int(msg.GetDataLen())
	if uint(size) > r.config.AppConf.MaxPackSize || len(msg.GetData()) > size {
		return false
	}
	body := make([]byte, size)
	copy(body, msg.GetData())
	if _, err := io.ReadFull(conn, body[len(msg.GetData()):]); err != nil {
		return false
	}
	msg.SetData(body)
	handedOff = submitReadTask(c, func() { defer PutMessage(msg); readerTaskHandler(c, msg) })
	return handedOff
}

func writeStream(conn net.Conn, data []byte) bool {
	for len(data) > 0 {
		n, err := conn.Write(data)
		if err != nil || n <= 0 {
			return false
		}
		data = data[n:]
	}
	return true
}
