package nets

import (
	"fmt"
	"github.com/451008604/nets/internal"
	"github.com/gorilla/websocket"
	"google.golang.org/protobuf/proto"
	"sync/atomic"
	"testing"
	"time"
)

func TestGetServerWS(t *testing.T) {
	initTest()

	// ====================== 启动服务 ======================
	go GetInstanceServerManager().RegisterServer(GetServerWS())
	// 等待服务启动
	time.Sleep(time.Second)

	// ====================== 注册路由 ======================
	instanceMsgHandler.AddRouter(int32(internal.Test_MsgId_Test_Echo), func() proto.Message { return &internal.Test_EchoRequest{} }, func(conn IConnection, message proto.Message) {
		req, ok := message.(*internal.Test_EchoRequest)
		if !ok || req == nil {
			return
		}
		res := &internal.Test_EchoResponse{Message: req.Message}
		conn.SendMsg(int32(internal.Test_MsgId_Test_Echo), res)
	})

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
		conn.Close()
	}

	if flagReceive != int32(connNum) {
		t.Error("TestGetServerWS", flagReceive)
	}
}
