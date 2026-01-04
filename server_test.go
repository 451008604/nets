package nets

import (
	"context"
	"github.com/451008604/nets/internal"
	"google.golang.org/protobuf/proto"
	"net/http"
	"runtime"
	"testing"
	"time"
)

// info periodically logs runtime stats until ctx is canceled.
func info(ctx context.Context, t *testing.T) {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)

			t.Logf(
				"SYS_CPU_NUM:%v\tALLOC:%v MB\tHEAP_ALLOC:%v MB\tTOTAL_ALLOC:%v MB\tCGO_CALL_NUM:%v\tGOROUTINE_NUM:%v",
				runtime.NumCPU(),
				memStats.Alloc/1024/1024,
				memStats.HeapAlloc/1024/1024,
				memStats.TotalAlloc/1024/1024,
				runtime.NumCgoCall(),
				runtime.NumGoroutine(),
			)
		}
	}
}

func Test_Server(t *testing.T) {
	go StartServer(t)

	time.Sleep(5 * time.Second)

	go Test_Client_WS(t)
	go Test_Client_KCP(t)

	time.Sleep(5 * time.Second)
}

func StartServer(t *testing.T) {
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
		reader := ConnPropertyGet[*http.Request](conn, ConnPropertyHttpReader)
		writer := ConnPropertyGet[http.ResponseWriter](conn, ConnPropertyHttpWriter)
		if reader == nil || writer == nil {
			return
		}
		msgReq, ok := message.(*Message)
		if !ok || msgReq == nil {
			return
		}
		conn.SendMsg(int32(internal.Test_MsgId_Test_None), msgReq)
	})

	// 周期打印运行时信息，测试结束自动退出
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go info(ctx, t)

	// 注册服务
	GetInstanceServerManager().RegisterServer(
		GetServerTCP(),
		GetServerWS(),
		GetServerHTTP(),
		GetServerKCP(),
	)
}
