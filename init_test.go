package nets

import (
	"encoding/json"
	"fmt"
	"github.com/451008604/nets/internal"
	"github.com/gorilla/websocket"
	"github.com/xtaci/kcp-go"
	"google.golang.org/protobuf/proto"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
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
	SetCustomServer(&CustomServer{AppConf: &AppConf{ConnRWTimeOut: 30}})
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
	})

	// ====================== 启动服务 ======================
	go GetInstanceServerManager().RegisterServer(GetServerHTTP(), GetServerKCP(), GetServerTCP(), GetServerWS())
	// 等待服务启动
	time.Sleep(time.Second * 3)

	code := m.Run()
	os.Exit(code)
}

func TestServer(t *testing.T) {
	connNum := 1
	for i := 0; i < connNum; i++ {
		// ====================================================== 发送Restful API 模式HTTP请求
		reqData := "testpoint"
		request, _ := http.NewRequest(http.MethodPost, fmt.Sprintf(`http://127.0.0.1:%v/testpoint`, defaultServer.AppConf.ServerHTTP.Port), strings.NewReader(reqData))
		resp, err := http.DefaultClient.Do(request)
		if err != nil {
			t.Fatal(err)
		}
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if string(body) != `{"msg_id":0,"data":"`+reqData+`"}` {
			t.Fatal("TestGetServerHTTP", string(body))
		}

		// ====================================================== 发送路由模式HTTP请求
		marshal, _ := json.Marshal(NewMsgPackage(int32(internal.Test_MsgId_Test_Echo), msgStr)) // {"msg_id":1001,"data":"{\"Message\":\"hello world\"}"}
		request, _ = http.NewRequest(http.MethodPost, fmt.Sprintf(`http://127.0.0.1:%v/`, defaultServer.AppConf.ServerHTTP.Port), strings.NewReader(string(marshal)))
		client2 := &http.Client{}
		resp, _ = client2.Do(request)
		body, _ = io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if string(body) != string(msgStr) {
			t.Fatal("TestGetServerHTTP", string(body))
		}

		// ====================================================== 测试WS
		connWs, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://127.0.0.1:%v", defaultServer.AppConf.ServerWS.Port), nil)
		if err != nil {
			t.Fatal(err)
		}
		// 发送消息
		_ = connWs.WriteMessage(websocket.BinaryMessage, defaultServer.DataPack.Pack(NewMsgPackage(int32(internal.Test_MsgId_Test_Echo), msgStr)))
		// 接收消息
		if _, message, _ := connWs.ReadMessage(); len(message) != 0 {
			if pack := NewDataPack().UnPack(message); pack != nil {
				if string(pack.GetData()) != string(msgStr) {
					t.Fatal("TestGetServerWS1", string(pack.GetData()))
				}
			}
		}
		_ = connWs.Close()

		// ====================================================== 测试TCP
		connTcp, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%v", defaultServer.AppConf.ServerTCP.Port))
		if err != nil {
			t.Fatal(err)
		}
		// 发送消息
		_, _ = connTcp.Write(defaultServer.DataPack.Pack(NewMsgPackage(int32(internal.Test_MsgId_Test_Echo), msgStr)))
		// 接收消息
		tcpBuf := make([]byte, 4096)
		if message, _ := connTcp.Read(tcpBuf); message != 0 {
			if pack := defaultServer.DataPack.UnPack(tcpBuf[:message]); pack != nil {
				if string(pack.GetData()) != string(msgStr) {
					t.Fatal("TestGetServerTCP1", string(pack.GetData()))
				}
			}
		}
		_ = connTcp.Close()

		// ====================================================== 测试KCP
		connKcp, err := kcp.DialWithOptions(fmt.Sprintf("127.0.0.1:%v", defaultServer.AppConf.ServerKCP.Port), nil, 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		// connKcp.SetNoDelay(1, 10, 2, 1)
		connKcp.SetNoDelay(1, 10, 2, 1) // nodelay, interval, resend, nc
		connKcp.SetStreamMode(true)
		connKcp.SetWindowSize(128, 128)
		// 发送消息
		_, _ = connKcp.Write(defaultServer.DataPack.Pack(NewMsgPackage(int32(internal.Test_MsgId_Test_Echo), msgStr)))
		// 接收消息
		// _ = connKcp.SetReadDeadline(time.Now().Add(time.Second))
		kcpBuf := make([]byte, 4096)
		if message, _ := connKcp.Read(kcpBuf); message != 0 {
			if pack := defaultServer.DataPack.UnPack(kcpBuf[:message]); pack != nil {
				if string(pack.GetData()) != string(msgStr) {
					t.Fatal("TestGetServerKCP1", string(pack.GetData()))
				}
			}
		}
		_ = connKcp.Close()
	}
}
