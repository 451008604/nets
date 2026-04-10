package nets

import (
	"fmt"
	"github.com/451008604/nets/internal"
	"net"
	"sync/atomic"
	"testing"
)

func TestGetServerTCP(t *testing.T) {
	// ====================== 发送请求 ======================
	connNum := 1000
	for i := 0; i < connNum; i++ {
		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%v", defaultServer.AppConf.ServerTCP.Port))
		if err != nil {
			t.Error(err)
			continue
		}
		// 发送消息
		_, _ = conn.Write(defaultServer.DataPack.Pack(NewMsgPackage(int32(internal.Test_MsgId_Test_Echo), msgStr)))

		// 接收消息
		buf := make([]byte, 4096)
		if message, _ := conn.Read(buf); message != 0 {
			if pack := defaultServer.DataPack.UnPack(buf[:message]); pack != nil {
				atomic.AddInt32(&flagReceive, 1)
			}
		}
		_ = conn.Close()
	}

	if flagReceive != int32(connNum) {
		t.Error("TestGetServerTCP", flagReceive)
	}
}
