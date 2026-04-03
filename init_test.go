package nets

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/451008604/nets/internal"
	"sync"
	"sync/atomic"
	"time"
)

var msgStr, _ = json.Marshal(&internal.Test_EchoRequest{Message: "hello world"})

var flagSend, flagReceive, flagOpened, flagClosed, flagErrCapture int32

func initTest() {
	// 重置标志位
	flagSend, flagReceive, flagOpened, flagClosed, flagErrCapture = 0, 0, 0, 0, 0
}

// ====================== 模拟测试连接 ======================
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
