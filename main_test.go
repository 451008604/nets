package nets

import (
	"context"
	"fmt"
	"github.com/451008604/nets/internal"
	"google.golang.org/protobuf/proto"
	"net/http"
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

// 打点标记测试逻辑触发次数
var flag1, flag2, flag3 int32

func Test_Server(t *testing.T) {
	msgHandler := GetInstanceMsgHandler()
	msgHandler.AddRouter(int32(internal.Test_MsgId_Test_Echo), func() proto.Message { return &internal.Test_EchoRequest{} }, echoRequest)
	msgHandler.AddRouter(int32(internal.Test_MsgId_Test_None), func() proto.Message { return &Message{} }, httpRequest)

	connManager := GetInstanceConnManager()
	connManager.SetConnOnOpened(func(conn IConnection) { atomic.AddInt32(&flag1, 1) })
	connManager.SetConnOnClosed(func(conn IConnection) { atomic.AddInt32(&flag2, 1) })

	for i := 0; i < 40000; i++ {
		wg.Add(1)
		connManager.Add(NewConnectionTest())
	}

	wg.Wait()
	time.Sleep(time.Second * 2)
	t.Log(flag1, flag2, flag3)
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
