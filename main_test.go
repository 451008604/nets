package nets

import (
	"context"
	"fmt"
	"github.com/451008604/nets/internal"
	"google.golang.org/protobuf/proto"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type connectionTest struct {
	*ConnectionBase
}

func NewConnectionTest() IConnection {
	c := &connectionTest{
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
	return fmt.Sprintf("%v", atomic.AddInt32(&id, 1))
}

var id = int32(0)
var wg = sync.WaitGroup{}

func Test_Server(t *testing.T) {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	printMem("start", &m)
	// ===========消息处理器===========
	msgHandler := GetInstanceMsgHandler()
	msgHandler.AddRouter(int32(internal.Test_MsgId_Test_Echo), func() proto.Message { return &internal.Test_EchoRequest{} }, echoRequest)
	msgHandler.AddRouter(int32(internal.Test_MsgId_Test_None), func() proto.Message { return &Message{} }, httpRequest)

	connManager := GetInstanceConnManager()
	runtime.GC()
	runtime.ReadMemStats(&m)
	printMem("init", &m)

	wg.Add(1)
	connManager.Add(NewConnectionTest())
	time.Sleep(time.Second * 2)
	runtime.GC()
	runtime.ReadMemStats(&m)
	printMem("end", &m)

	for j := 0; j < 4; j++ {
		for i := 0; i < 10000; i++ {
			wg.Add(1)
			connManager.Add(NewConnectionTest())
		}
		runtime.GC()
		time.Sleep(time.Second * 2)
		runtime.ReadMemStats(&m)
		printMem("end", &m)
	}

	wg.Wait()
	time.Sleep(time.Second * 20)
	runtime.GC()
	runtime.ReadMemStats(&m)
	printMem("end", &m)
}

// start: Alloc = 328 KiB, TotalAlloc = 328 KiB, Sys = 7826 KiB, NumGC = 0, NumGoroutine = 2
// init: Alloc = 1905 KiB, TotalAlloc = 1918 KiB, Sys = 13266 KiB, NumGC = 1, NumGoroutine = 3
// end: Alloc = 23675 KiB, TotalAlloc = 247315 KiB, Sys = 562838 KiB, NumGC = 10, NumGoroutine = 5

// start: Alloc = 328 KiB, TotalAlloc = 328 KiB, Sys = 7826 KiB, NumGC = 0, NumGoroutine = 2
// init: Alloc = 1900 KiB, TotalAlloc = 1913 KiB, Sys = 8658 KiB, NumGC = 1, NumGoroutine = 3
// end: Alloc = 24195 KiB, TotalAlloc = 108878 KiB, Sys = 427882 KiB, NumGC = 9, NumGoroutine = 5

// start: Alloc = 328 KiB, TotalAlloc = 328 KiB, Sys = 7826 KiB, NumGC = 0, NumGoroutine = 2
// init: Alloc = 1905 KiB, TotalAlloc = 1918 KiB, Sys = 13266 KiB, NumGC = 1, NumGoroutine = 3
// end: Alloc = 24327 KiB, TotalAlloc = 108355 KiB, Sys = 465386 KiB, NumGC = 9, NumGoroutine = 5

// start: Alloc = 328 KiB, TotalAlloc = 328 KiB, Sys = 7826 KiB, NumGC = 0, NumGoroutine = 2
// init: Alloc = 1905 KiB, TotalAlloc = 1918 KiB, Sys = 13010 KiB, NumGC = 1, NumGoroutine = 3
// end: Alloc = 24088 KiB, TotalAlloc = 248069 KiB, Sys = 578966 KiB, NumGC = 9, NumGoroutine = 5

// start: Alloc = 328 KiB, TotalAlloc = 328 KiB, Sys = 7826 KiB, NumGC = 0, NumGoroutine = 2
// init: Alloc = 1905 KiB, TotalAlloc = 1918 KiB, Sys = 13010 KiB, NumGC = 1, NumGoroutine = 3
// end: Alloc = 23697 KiB, TotalAlloc = 212447 KiB, Sys = 520726 KiB, NumGC = 10, NumGoroutine = 5

// start: Alloc = 328 KiB, TotalAlloc = 328 KiB, Sys = 7826 KiB, NumGC = 0, NumGoroutine = 2
// init: Alloc = 1905 KiB, TotalAlloc = 1918 KiB, Sys = 13010 KiB, NumGC = 1, NumGoroutine = 3
// end: Alloc = 23977 KiB, TotalAlloc = 143668 KiB, Sys = 462378 KiB, NumGC = 10, NumGoroutine = 5

func printMem(tag string, ms *runtime.MemStats) {
	fmt.Printf("%s: Alloc = %d KiB, TotalAlloc = %d KiB, Sys = %d KiB, NumGC = %d, NumGoroutine = %d\n",
		tag, ms.Alloc/1024, ms.TotalAlloc/1024, ms.Sys/1024, ms.NumGC, runtime.NumGoroutine())
}

func echoRequest(conn IConnection, message proto.Message) {
	req, ok := message.(*internal.Test_EchoRequest)
	if !ok || req == nil {
		return
	}
	res := &internal.Test_EchoResponse{Message: req.Message}
	conn.SendMsg(int32(internal.Test_MsgId_Test_Echo), res)
}

func httpRequest(conn IConnection, message proto.Message) {
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
}
