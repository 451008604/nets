package nets

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/451008604/nets/internal"
	"google.golang.org/protobuf/proto"
)

type ConnectionTest struct {
	*ConnectionBase
}

func NewConnectionTest() *ConnectionTest {
	c := &ConnectionTest{
		ConnectionBase: &ConnectionBase{
			connId:        fmt.Sprintf("%X-%v", time.Now().Unix(), atomic.AddUint32(&connIdSeed, 1)),
			msgBuffChan:   make(chan []byte, defaultServer.AppConf.MaxMsgChanLen),
			taskQueue:     make(chan func(), defaultServer.AppConf.WorkerTaskMaxLen),
			property:      map[string]any{},
			propertyMutex: sync.RWMutex{},
		},
	}
	c.exitCtx, c.exitCtxCancel = context.WithCancel(context.Background())
	c.ConnectionBase.conn = c
	return c
}

func (c *ConnectionTest) StartReader() bool {
	if c.IsClose() {
		return false
	}
	if msgReq, ok := c.GetProperty("msgReq").([]byte); ok {
		c.RemoveProperty("msgReq")
		atomic.AddInt32(&flagSend, 1)
		// 封装请求数据传入处理函数
		c.DoTask(func() {
			readerTaskHandler(c, defaultServer.DataPack.UnPack(msgReq))
		})
	}
	time.Sleep(time.Millisecond * 1)
	return true
}

func (c *ConnectionTest) StartWriter(data []byte) bool {
	c.SetProperty("msgRes", data)
	return true
}

func (c *ConnectionTest) RemoteAddrStr() string {
	return ""
}

var msgStr = []byte(`{"Message":"hello world"}`)

var flagSend, flagReceive, flagOpened, flagClosed, flagErrCapture, flagFilter int32

func initTest() {
	instanceMsgHandlerOnce = sync.Once{}
	instanceConnManagerOnce = sync.Once{}
	instanceMsgHandler = GetInstanceMsgHandler()
	instanceConnManager = GetInstanceConnManager()
	flagSend, flagReceive, flagOpened, flagClosed, flagErrCapture, flagFilter = int32(0), int32(0), int32(0), int32(0), int32(0), int32(0)
}

// 测试1百万个连接收发消息
func TestConnections(t *testing.T) {
	initTest()

	instanceConnManager.SetConnOnOpened(func(conn IConnection) { atomic.AddInt32(&flagOpened, 1) })
	instanceConnManager.SetConnOnClosed(func(conn IConnection) { atomic.AddInt32(&flagClosed, 1) })
	instanceMsgHandler.AddRouter(int32(internal.Test_MsgId_Test_Echo), func() proto.Message { return &internal.Test_EchoRequest{} }, func(conn IConnection, message proto.Message) {
		req, ok := message.(*internal.Test_EchoRequest)
		if !ok || req == nil {
			return
		}
		res := &internal.Test_EchoResponse{Message: req.Message}
		conn.SendMsg(int32(internal.Test_MsgId_Test_Echo), res)
	})

	var cCount = 10000
	var wg = sync.WaitGroup{}

	for i := 0; i < cCount; i++ {
		wg.Add(1)
		connection := NewConnectionTest()
		instanceConnManager.Add(connection)
		// 通过设置属性模拟数据传入
		connection.SetProperty("msgReq", defaultServer.DataPack.Pack(NewMsgPackage(int32(internal.Test_MsgId_Test_Echo), msgStr)))
		go func() {
			for {
				if msgRes, ok := connection.GetProperty("msgRes").([]byte); ok {
					connection.RemoveProperty("msgRes")
					if string(defaultServer.DataPack.UnPack(msgRes).GetData()) == string(msgStr) {
						atomic.AddInt32(&flagReceive, 1)
					}
					instanceConnManager.Remove(connection)
					wg.Done()
					break
				}
				time.Sleep(time.Millisecond * 1)
			}
		}()
	}
	wg.Wait()
	time.Sleep(time.Second)
	if flagSend != int32(cCount) || flagReceive != int32(cCount) || flagOpened != int32(cCount) || flagClosed != int32(cCount) {
		t.Error("flagSend", flagSend, "flagReceive", flagReceive, "flagOpened", flagOpened, "flagClosed", flagClosed)
		return
	}
}

// 测试 taskFun 发生panic 时异常捕获
func TestMsgHandler_SetErrCapture(t *testing.T) {
	initTest()

	instanceMsgHandler.SetErrCapture(func(conn IConnection, panicInfo string) {
		atomic.AddInt32(&flagErrCapture, 1)
	})

	connection := NewConnectionTest()
	instanceConnManager.Add(connection)

	// panicInfo Test_MsgId_Test_Echo panic
	instanceMsgHandler.AddRouter(int32(internal.Test_MsgId_Test_Echo), func() proto.Message { return &internal.Test_EchoRequest{} }, func(conn IConnection, message proto.Message) {
		panic("Test_MsgId_Test_Echo panic")
	})
	connection.SetProperty("msgReq", defaultServer.DataPack.Pack(NewMsgPackage(int32(internal.Test_MsgId_Test_Echo), msgStr)))
	// panicInfo runtime error: integer divide by zero
	connection.DoTask(func() {
		n := int32(0)
		_ = 3 / int32(n)
	})
	// panicInfo runtime error: invalid memory address or nil pointer dereference
	connection.DoTask(func() {
		type testStruct struct {
			n int32
		}
		var test *testStruct
		test.n = 10
	})

	time.Sleep(time.Second)
	if flagErrCapture != int32(3) {
		t.Error("TestMsgHandler_SetErrCapture", flagErrCapture)
		return
	}
}

// 测试 http 请求兼容性
func TestHttpConnection(t *testing.T) {
	initTest()

	// Restful API 模式
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

	// 消息ID路由模式
	instanceMsgHandler.AddRouter(int32(internal.Test_MsgId_Test_Echo), func() proto.Message { return &internal.Test_EchoRequest{} }, func(conn IConnection, message proto.Message) {
		req, ok := message.(*internal.Test_EchoRequest)
		if !ok || req == nil {
			return
		}
		res := &internal.Test_EchoResponse{Message: req.Message}
		conn.SendMsg(int32(internal.Test_MsgId_Test_Echo), res)
	})

	// 启动服务
	go GetInstanceServerManager().RegisterServer(GetServerHTTP())

	time.Sleep(time.Second)

	// 发送Restful API 模式请求
	reqData := "testpoint"
	request, _ := http.NewRequest(http.MethodPost, `http://127.0.0.1:`+defaultServer.AppConf.ServerHTTP.Port+`/testpoint`, strings.NewReader(reqData))
	resp, _ := http.DefaultClient.Do(request)
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if string(body) != `{"msg_id":0,"data":"`+reqData+`"}` {
		t.Error("TestHttpConnection", string(body))
	}

	// 发送路由模式请求
	httpReq := NewMsgPackage(int32(internal.Test_MsgId_Test_Echo), msgStr)
	marshal, _ := json.Marshal(httpReq) // {"msg_id":1001,"data":"{\"Message\":\"hello world\"}"}
	request, _ = http.NewRequest(http.MethodPost, `http://127.0.0.1:`+defaultServer.AppConf.ServerHTTP.Port+`/`, strings.NewReader(string(marshal)))
	resp, _ = http.DefaultClient.Do(request)
	body, _ = io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if string(body) != string(msgStr) {
		t.Error("TestHttpConnection", string(body))
	}
}

func TestMsgHandler_SetFilter(t *testing.T) {
	initTest()

	instanceMsgHandler.SetFilter(func(conn IConnection, msg IMessage) bool {
		conn.SetProperty("filterKey", "filterValue")
		return true
	})

	connection := NewConnectionTest()
	instanceConnManager.Add(connection)
	instanceMsgHandler.AddRouter(int32(internal.Test_MsgId_Test_Echo), func() proto.Message { return &internal.Test_EchoRequest{} }, func(conn IConnection, message proto.Message) {
		if v, ok := conn.GetProperty("filterKey").(string); ok {
			if v != "filterValue" {
				t.Error("TestMsgHandler_SetFilter", v)
			}
		}
	})
	connection.SetProperty("msgReq", defaultServer.DataPack.Pack(NewMsgPackage(int32(internal.Test_MsgId_Test_Echo), msgStr)))
}
