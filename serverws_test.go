package nets

import (
	"fmt"
	"github.com/451008604/nets/internal"
	"github.com/gorilla/websocket"
	"sync/atomic"
	"testing"
)

func TestGetServerWS(t *testing.T) {
	// ====================== 发送请求 ======================
	connNum := 1000
	for i := 0; i < connNum; i++ {
		conn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://127.0.0.1:%v", defaultServer.AppConf.ServerWS.Port), nil)
		if err != nil {
			t.Error(err)
			continue
		}

		// 发送消息
		_ = conn.WriteMessage(websocket.BinaryMessage, defaultServer.DataPack.Pack(NewMsgPackage(int32(internal.Test_MsgId_Test_Echo), msgStr)))

		// 接收消息
		if _, message, _ := conn.ReadMessage(); len(message) != 0 {
			if pack := NewDataPack().UnPack(message); pack != nil {
				atomic.AddInt32(&flagReceive, 1)
			}
		}
		_ = conn.Close()
	}

	if flagReceive != int32(connNum) {
		t.Error("TestGetServerWS", flagReceive)
	}
}
