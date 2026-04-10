package nets

import (
	"fmt"
	"github.com/451008604/nets/internal"
	"github.com/xtaci/kcp-go"
	"sync/atomic"
	"testing"
	"time"
)

func TestGetServerKCP(t *testing.T) {
	// ====================== 发送请求 ======================
	connNum := 1000
	for i := 0; i < connNum; i++ {
		conn, err := kcp.DialWithOptions(fmt.Sprintf("127.0.0.1:%v", defaultServer.AppConf.ServerKCP.Port), nil, 0, 0)
		if err != nil {
			t.Error(err)
			continue
		}
		// conn.SetNoDelay(1, 10, 2, 1)
		conn.SetNoDelay(1, 10, 2, 1) // nodelay, interval, resend, nc
		conn.SetStreamMode(true)
		conn.SetWindowSize(128, 128)

		// 发送消息
		_, _ = conn.Write(defaultServer.DataPack.Pack(NewMsgPackage(int32(internal.Test_MsgId_Test_Echo), msgStr)))

		// 接收消息
		_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		buf := make([]byte, 4096)
		if message, _ := conn.Read(buf); message != 0 {
			if pack := defaultServer.DataPack.UnPack(buf[:message]); pack != nil {
				atomic.AddInt32(&flagReceive, 1)
			}
		}

		_ = conn.Close()
	}

	time.Sleep(time.Second)
	if flagReceive != int32(connNum) {
		t.Error("TestGetServerKCP", flagReceive)
	}
}
