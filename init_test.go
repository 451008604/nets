package nets

import (
	"encoding/json"
	"fmt"
	"github.com/451008604/nets/internal"
	"google.golang.org/protobuf/proto"
	"net/http"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

var msgStr, _ = json.Marshal(&internal.Test_EchoRequest{Message: "hello world"})

type testFlag struct {
	flagSend       int32
	flagReceive    int32
	flagOpened     int32
	flagClosed     int32
	flagErrCapture int32
}

var flag = &testFlag{}

func TestMain(m *testing.M) {
	GetInstanceConnManager().SetConnOpened(func(conn IConnection) { atomic.AddInt32(&flag.flagOpened, 1) })
	GetInstanceConnManager().SetConnClosed(func(conn IConnection) { atomic.AddInt32(&flag.flagClosed, 1) })
	GetInstanceMsgHandler().SetFilter(func(conn IConnection, msg IMessage) bool {
		conn.SetProperty("filterKey", "filterValue")
		return true
	})
	GetInstanceMsgHandler().SetErrCapture(func(conn IConnection, panicInfo string) {
		atomic.AddInt32(&flag.flagErrCapture, 1)
	})
	// ====================== 注册路由 ======================
	// Restful API 模式
	GetInstanceMsgHandler().AddRouter(int32(internal.Test_MsgId_Test_None), func() proto.Message { return &Message{} }, func(conn IConnection, message proto.Message) {
		reader := conn.GetProperty(ConnPropertyHttpReader).(*http.Request)
		writer := conn.GetProperty(ConnPropertyHttpWriter).(http.ResponseWriter)
		if reader == nil || writer == nil {
			return
		}
		msgReq, ok := message.(*Message)
		if !ok || msgReq == nil {
			return
		}

		// t.Log("Method", reader.Method, "RequestURI", reader.RequestURI, "Data", string(msgReq.GetData()))
		conn.SendMsg(int32(internal.Test_MsgId_Test_None), msgReq)
		atomic.AddInt32(&flag.flagReceive, 1)
	})
	// 消息ID 路由模式
	GetInstanceMsgHandler().AddRouter(int32(internal.Test_MsgId_Test_Echo), func() proto.Message { return &internal.Test_EchoRequest{} }, func(conn IConnection, message proto.Message) {
		req, ok := message.(*internal.Test_EchoRequest)
		if !ok || req == nil {
			return
		}
		res := &internal.Test_EchoResponse{Message: req.Message}
		conn.SendMsg(int32(internal.Test_MsgId_Test_Echo), res)
		atomic.AddInt32(&flag.flagReceive, 1)

		if conn.RemoteAddrStr() == "" {
			fmt.Println("conn.RemoteAddrStr() is empty")
			return
		}

		if v, ok := conn.GetProperty("filterKey").(string); ok {
			if v != "filterValue" {
				fmt.Println("TestMsgHandler_SetFilter", v)
			}
		}

		conn.Close()
		conn.SendMsg(int32(internal.Test_MsgId_Test_Echo), res)

		panic("Test_MsgId_Test_Echo panic")
	})

	// ====================== 启动服务 ======================
	go GetInstanceServerManager().RegisterServer(GetServerHTTP(), GetServerKCP(), GetServerTCP(), GetServerWS())
	// 等待服务启动
	time.Sleep(time.Second * 5)

	code := m.Run()
	os.Exit(code)
}
