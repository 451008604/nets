package nets

import (
	"context"
	"fmt"
	"github.com/451008604/nets/internal"
	"google.golang.org/protobuf/proto"
	"sync"
	"sync/atomic"
	"testing"
	"time"
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

	var cCount = 1000000
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
