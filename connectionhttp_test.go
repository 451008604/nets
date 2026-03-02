package nets

import (
	"fmt"
	"github.com/451008604/nets/internal"
	"google.golang.org/protobuf/proto"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func Test_connectionHTTP_StartWriter(t *testing.T) {
	initTest()

	// ====================== 启动服务 ======================
	go GetInstanceServerManager().RegisterServer(GetServerHTTP())
	// 等待服务启动
	time.Sleep(time.Second)

	// ====================== 注册路由 ======================
	// Restful API 模式
	instanceMsgHandler.SetFilter(func(conn IConnection, msg IMessage) bool {
		return true
	})
	instanceMsgHandler.AddRouter(int32(internal.Test_MsgId_Test_None), func() proto.Message { return &Message{} }, func(conn IConnection, message proto.Message) {
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
}
