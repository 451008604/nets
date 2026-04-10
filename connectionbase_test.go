package nets

import (
	"github.com/451008604/nets/internal"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// 测试1百万个连接收发消息
func TestConnections(t *testing.T) {
	var cCount = 1000000
	var wg = sync.WaitGroup{}

	for i := 0; i < cCount; i++ {
		wg.Add(1)
		connection := NewConnectionTest()
		GetInstanceConnManager().Add(connection)
		// 通过设置属性模拟数据传入
		connection.SetProperty("msgReq", defaultServer.DataPack.Pack(NewMsgPackage(int32(internal.Test_MsgId_Test_Echo), msgStr)))
		go func() {
			for {
				if msgRes, ok := connection.GetProperty("msgRes").([]byte); ok {
					connection.RemoveProperty("msgRes")
					if string(defaultServer.DataPack.UnPack(msgRes).GetData()) == string(msgStr) {
						atomic.AddInt32(&flagReceive, 1)
					}
					GetInstanceConnManager().Remove(connection)
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
