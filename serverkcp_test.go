package nets

import (
	"fmt"
	"github.com/451008604/nets/internal"
	"github.com/xtaci/kcp-go"
	"testing"
	"time"
)

func TestGetServerKCP(t *testing.T) {
	connNum := 10000
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
		// _ = conn.SetReadDeadline(time.Now().Add(time.Second))
		buf := make([]byte, 4096)
		if message, _ := conn.Read(buf); message != 0 {
			if pack := defaultServer.DataPack.UnPack(buf[:message]); pack != nil {
				if string(pack.GetData()) != string(msgStr) {
					t.Error("TestGetServerKCP1", string(pack.GetData()))
				}
			}
		}

		_ = conn.Close()
	}

	if flag.flagReceive != int32(connNum) {
		t.Error("TestGetServerKCP2", flag.flagReceive)
	}

	t.Cleanup(func() {
		time.Sleep(time.Second * 3)
		fmt.Printf("%+v\n", flag)
	})
}
