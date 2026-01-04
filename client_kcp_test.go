package nets

import (
	"github.com/xtaci/kcp-go"
	"net"
	"sync"
	"testing"
)

/*
测试1
*/
func ClientKcp(t *testing.T, wg *sync.WaitGroup, data []byte) {
	conn, err := kcp.DialWithOptions("127.0.0.1:17004", nil, 0, 0)
	if err != nil {
		t.Error(err)
		return
	}
	defer func(conn *kcp.UDPSession) {
		_ = conn.Close()
		wg.Done()
	}(conn)
	// conn.SetNoDelay(1, 10, 2, 1)

	go func(c net.Conn) {
		buf := make([]byte, 4096)
		if message, _ := c.Read(buf); message != 0 {
			if pack := NewDataPack().UnPack(buf[:message]); pack != nil {
				// t.Logf("服务器：%v - %s\n", pack.GetMsgId(), pack.GetData())
			}
		}
	}(conn)

	// 发送消息
	_, _ = conn.Write(append(append(append(append(data, data...), data...), data...), data...))
}
