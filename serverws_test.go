package nets

import (
	"fmt"
	"github.com/451008604/nets/internal"
	"github.com/gorilla/websocket"
	"testing"
	"time"
)

func TestGetServerWS(t *testing.T) {
	connNum := 10000
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
				if string(pack.GetData()) != string(msgStr) {
					t.Error("TestGetServerWS1", string(pack.GetData()))
				}
			}
		}
		_ = conn.Close()
	}

	if flag.flagReceive != int32(connNum) {
		t.Error("TestGetServerWS2", flag.flagReceive)
	}
	t.Cleanup(func() {
		time.Sleep(time.Second * 3)
		fmt.Printf("%+v\n", flag)
	})
}
