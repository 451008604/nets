package nets

import (
	"context"
	"fmt"
	"net/http"
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

var flagSend, flagReceive = int32(0), int32(0)
var flagOpened, flagClosed, flagErrCapture, flagFilter = int32(0), int32(0), int32(0), int32(0)

func Test_Server(t *testing.T) {
	msgHandler := GetInstanceMsgHandler()
	msgHandler.SetErrCapture(errCapture)
	msgHandler.SetFilter(filter)
	msgHandler.AddRouter(int32(internal.Test_MsgId_Test_Echo), func() proto.Message { return &internal.Test_EchoRequest{} }, echoRequest)
	msgHandler.AddRouter(int32(internal.Test_MsgId_Test_None), func() proto.Message { return &Message{} }, httpRequest)
	connManager := GetInstanceConnManager()

	// 测试1百万个连接收发消息
	t.Run("Test 1,000,000 connections for data transmission and reception.", func(t *testing.T) {
		var cCount = 1000000
		var wg = sync.WaitGroup{}

		connManager.SetConnOnOpened(func(conn IConnection) {
			atomic.AddInt32(&flagOpened, 1)
		})
		connManager.SetConnOnClosed(func(conn IConnection) {
			atomic.AddInt32(&flagClosed, 1)
		})

		for i := 0; i < cCount; i++ {
			wg.Add(1)
			msgStr := `{"Message":"hello world"}`
			connection := NewConnectionTest()
			connManager.Add(connection)
			connection.SetProperty("msgReq", defaultServer.DataPack.Pack(NewMsgPackage(int32(internal.Test_MsgId_Test_Echo), []byte(msgStr))))
			go func() {
				for {
					if msgRes, ok := connection.GetProperty("msgRes").([]byte); ok {
						connection.RemoveProperty("msgRes")
						if string(defaultServer.DataPack.UnPack(msgRes).GetData()) == msgStr {
							atomic.AddInt32(&flagReceive, 1)
						}
						connManager.Remove(connection)
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
		t.Log("connections", cCount, "Successfully")
	})

	// 测试 taskFun 发生panic 时异常捕获
	t.Run("Handle unexpected errors during the \"taskFun\" test.", func(t *testing.T) {

	})
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

func errCapture(conn IConnection, taskFun func(), panicInfo string) {
	atomic.AddInt32(&flagErrCapture, 1)
}

func filter(conn IConnection, msg IMessage) bool {
	atomic.AddInt32(&flagFilter, 1)
	return true
}
