package nets

import (
	"encoding/json"
	"fmt"
	"github.com/451008604/nets/internal"
	"google.golang.org/protobuf/proto"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

// 测试 http 请求兼容性
func TestGetServerHTTP(t *testing.T) {
	initTest()

	// ====================== 启动服务 ======================
	go GetInstanceServerManager().RegisterServer(GetServerHTTP())
	// 等待服务启动
	time.Sleep(time.Second)

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

		if conn.RemoteAddrStr() == "" {
			t.Error("conn.RemoteAddrStr() is empty")
			return
		}
		conn.Close()
		conn.SendMsg(int32(internal.Test_MsgId_Test_Echo), msgReq)
	})

	// 消息ID 路由模式
	GetInstanceMsgHandler().AddRouter(int32(internal.Test_MsgId_Test_Echo), func() proto.Message { return &internal.Test_EchoRequest{} }, func(conn IConnection, message proto.Message) {
		req, ok := message.(*internal.Test_EchoRequest)
		if !ok || req == nil {
			return
		}

		res := &internal.Test_EchoResponse{Message: req.Message}
		conn.SendMsg(int32(internal.Test_MsgId_Test_Echo), res)

		if conn.RemoteAddrStr() == "" {
			t.Error("conn.RemoteAddrStr() is empty")
			return
		}
		conn.Close()
		conn.SendMsg(int32(internal.Test_MsgId_Test_Echo), res)
	})

	// ====================== 发送请求 ======================
	// 发送Restful API 模式请求
	reqData := "testpoint"
	request, _ := http.NewRequest(http.MethodPost, fmt.Sprintf(`http://127.0.0.1:%v/testpoint`, defaultServer.AppConf.ServerHTTP.Port), strings.NewReader(reqData))
	resp, _ := http.DefaultClient.Do(request)
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if string(body) != `{"msg_id":0,"data":"`+reqData+`"}` {
		t.Error("TestGetServerHTTP", string(body))
	}

	// 发送路由模式请求
	marshal, _ := json.Marshal(NewMsgPackage(int32(internal.Test_MsgId_Test_Echo), msgStr)) // {"msg_id":1001,"data":"{\"Message\":\"hello world\"}"}
	request, _ = http.NewRequest(http.MethodPost, fmt.Sprintf(`http://127.0.0.1:%v/`, defaultServer.AppConf.ServerHTTP.Port), strings.NewReader(string(marshal)))
	resp, _ = http.DefaultClient.Do(request)
	body, _ = io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if string(body) != string(msgStr) {
		t.Error("TestGetServerHTTP", string(body))
	}
}
