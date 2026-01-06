package nets

import (
	"context"
	"fmt"
	"github.com/451008604/nets/internal"
	"google.golang.org/protobuf/proto"
	"net/http"
	"runtime"
	"sync"
	"testing"
	"time"
)

type connectionTest struct {
	*ConnectionBase
}

func NewConnectionTest() IConnection {
	c := &connectionTest{
		ConnectionBase: &ConnectionBase{
			connId:        GetInstanceConnManager().NewConnId(),
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

func (c *connectionTest) StartReader() bool {
	// 封装请求数据传入处理函数
	c.DoTask(func() {
		readerTaskHandler(c, NewMsgPackage(int32(internal.Test_MsgId_Test_Echo), []byte(`{"Message":"hello world"}`)))
	})

	// 这里等待1秒，模拟阻塞接收消息，否则会触发限流
	time.Sleep(time.Second)
	return true
}

func (c *connectionTest) StartWriter(data []byte) bool {

	wg.Done()
	return false
}

func (c *connectionTest) RemoteAddrStr() string {
	return ""
}

var wg = sync.WaitGroup{}

func Test_Server(t *testing.T) {
	// ===========消息处理器===========
	msgHandler := GetInstanceMsgHandler()
	// 添加路由：Echo
	msgHandler.AddRouter(int32(internal.Test_MsgId_Test_Echo), func() proto.Message { return &internal.Test_EchoRequest{} }, func(conn IConnection, message proto.Message) {
		req, ok := message.(*internal.Test_EchoRequest)
		if !ok || req == nil {
			return
		}
		res := &internal.Test_EchoResponse{Message: req.Message}
		conn.SendMsg(int32(internal.Test_MsgId_Test_Echo), res)
	})

	// 添加路由：None（HTTP透传示例）
	msgHandler.AddRouter(int32(internal.Test_MsgId_Test_None), func() proto.Message { return &Message{} }, func(conn IConnection, message proto.Message) {
		reader := conn.GetProperty(ConnPropertyHttpReader).(*http.Request)
		writer := conn.GetProperty(ConnPropertyHttpWriter).(http.ResponseWriter)
		if reader == nil || writer == nil {
			return
		}
		msgReq, ok := message.(*Message)
		if !ok || msgReq == nil {
			return
		}
		conn.SendMsg(int32(internal.Test_MsgId_Test_None), msgReq)
	})

	connManager := GetInstanceConnManager()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	printMem("start", &m)

	for i := 0; i < 10000; i++ {
		wg.Add(1)
		connManager.Add(NewConnectionTest())
	}

	wg.Wait()

	time.Sleep(time.Second * 2)
	runtime.GC()
	runtime.ReadMemStats(&m)
	printMem("end", &m)

}

func printMem(tag string, ms *runtime.MemStats) {
	fmt.Printf("%s: Alloc = %d KiB, TotalAlloc = %d KiB, Sys = %d KiB, NumGC = %d, NumGoroutine = %d\n",
		tag, ms.Alloc/1024, ms.TotalAlloc/1024, ms.Sys/1024, ms.NumGC, runtime.NumGoroutine())
}
