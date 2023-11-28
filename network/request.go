package network

import (
	"github.com/451008604/nets/iface"
)

type request struct {
	conn iface.IConnection // 已经和客户端建立好的连接
	msg  iface.IMessage    // 客户端请求的数据
}

func (r *request) GetConnection() iface.IConnection {
	return r.conn
}

func (r *request) GetData() []byte {
	return r.msg.GetData()
}

func (r *request) GetMsgID() int32 {
	return int32(r.msg.GetMsgId())
}
