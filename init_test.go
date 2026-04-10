package nets

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/451008604/nets/internal"
	"google.golang.org/protobuf/proto"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var msgStr, _ = json.Marshal(&internal.Test_EchoRequest{Message: "hello world"})

var flagSend, flagReceive, flagOpened, flagClosed, flagErrCapture int32

func TestMain(m *testing.M) {
	GetInstanceConnManager().SetConnOnOpened(func(conn IConnection) { atomic.AddInt32(&flagOpened, 1) })
	GetInstanceConnManager().SetConnOnClosed(func(conn IConnection) { atomic.AddInt32(&flagClosed, 1) })
	GetInstanceMsgHandler().SetFilter(func(conn IConnection, msg IMessage) bool {
		conn.SetProperty("filterKey", "filterValue")
		return true
	})
	// ====================== 注册路由 ======================
	GetInstanceMsgHandler().AddRouter(int32(internal.Test_MsgId_Test_Echo), func() proto.Message { return &internal.Test_EchoRequest{} }, func(conn IConnection, message proto.Message) {
		if v, ok := conn.GetProperty("filterKey").(string); ok {
			if v != "filterValue" {
				fmt.Println("TestMsgHandler_SetFilter", v)
			}
		}
		req, ok := message.(*internal.Test_EchoRequest)
		if !ok || req == nil {
			return
		}
		res := &internal.Test_EchoResponse{Message: req.Message}
		conn.SendMsg(int32(internal.Test_MsgId_Test_Echo), res)

		if conn.RemoteAddrStr() == "" {
			fmt.Println("conn.RemoteAddrStr() is empty")
			return
		}
		time.Sleep(time.Millisecond * 100)
		conn.Close()
		conn.SendMsg(int32(internal.Test_MsgId_Test_Echo), res)
	})

	// ====================== 启动服务 ======================
	go GetInstanceServerManager().RegisterServer(GetServerHTTP(), GetServerKCP(), GetServerTCP(), GetServerWS())
	// 等待服务启动
	time.Sleep(time.Second)

	code := m.Run()
	os.Exit(code)
}

// ============================================ 模拟连接 ============================================
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

// ============================================ 模拟连接 ============================================
