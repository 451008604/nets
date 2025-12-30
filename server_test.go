package nets

import (
	"fmt"
	"github.com/451008604/nets/internal"
	"google.golang.org/protobuf/proto"
	"net/http"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func info(t *testing.T) {
	for {
		select {
		case <-time.Tick(time.Second):
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			var mapping [][]any
			mapping = append(mapping, []any{"SYS_CPU_NUM", runtime.NumCPU()})
			mapping = append(mapping, []any{"ALLOC", fmt.Sprintf("%v MB", memStats.Alloc/1024/1024)})
			mapping = append(mapping, []any{"HEAP_ALLOC", fmt.Sprintf("%v MB", memStats.HeapAlloc/1024/1024)})
			mapping = append(mapping, []any{"TOTAL_ALLOC", fmt.Sprintf("%v MB", memStats.TotalAlloc/1024/1024)})
			mapping = append(mapping, []any{"CGO_CALL_NUM", runtime.NumCgoCall()})
			mapping = append(mapping, []any{"GOROUTINE_NUM", runtime.NumGoroutine()})

			str := ""
			for _, v := range mapping {
				str += fmt.Sprintf("%v：%v\t", v[0], v[1])
			}
			t.Log(str)
		}
	}
}

func Test_Server(t *testing.T) {
	tests := []struct {
		name string
		want *MsgHandler
	}{
		{
			name: "Test_Server",
			want: GetInstanceMsgHandler(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ===========消息处理器===========
			msgHandler := GetInstanceMsgHandler()
			if !reflect.DeepEqual(msgHandler, tt.want) {
				t.Errorf("GetInstanceMsgHandler() = %v, want %v", msgHandler, tt.want)
			}

			// 添加一个路由
			msgHandler.AddRouter(int32(internal.Test_MsgId_Test_Echo), func() proto.Message { return &internal.Test_EchoRequest{} }, func(conn IConnection, message proto.Message) {
				// do something ...
				req := message.(*internal.Test_EchoRequest)
				res := &internal.Test_EchoResponse{Message: req.Message}
				conn.SendMsg(int32(internal.Test_MsgId_Test_Echo), res)
			})
			msgHandler.AddRouter(int32(internal.Test_MsgId_Test_None), func() proto.Message { return &Message{} }, func(conn IConnection, message proto.Message) {
				reader := ConnPropertyGet[*http.Request](conn, ConnPropertyHttpReader)
				writer := ConnPropertyGet[http.ResponseWriter](conn, ConnPropertyHttpWriter)
				if reader == nil || writer == nil {
					return
				}
				msgReq := message.(*Message)
				conn.SendMsg(int32(internal.Test_MsgId_Test_None), msgReq)
			})

			go info(t)
			// 注册服务
			GetInstanceServerManager().RegisterServer(
				GetServerTCP(),
				GetServerWS(),
				GetServerHTTP(),
				GetServerKCP(),
			)
		})
	}
}
