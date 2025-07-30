package main

import (
	"fmt"
	"github.com/451008604/nets/iface"
	_ "github.com/451008604/nets/module"
	"github.com/451008604/nets/network"
	pb "github.com/451008604/nets/proto/bin"
	"google.golang.org/protobuf/proto"
	"net/http"
	"runtime"
)

// 服务指标监控
func listenChannelStatus() {
	serveMux := http.NewServeMux()
	server := &http.Server{Addr: ":17000", Handler: serveMux}
	serveMux.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(info()))
	})
	fmt.Printf("%v\n", server.ListenAndServe())
}

func info() string {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	var mapping [][]any
	mapping = append(mapping, []any{"GO_ROOT", runtime.GOROOT()})
	mapping = append(mapping, []any{"SYS_CPU_NUM", runtime.NumCPU()})
	mapping = append(mapping, []any{"ALLOC", fmt.Sprintf("%v MB", memStats.Alloc/1024/1024)})
	mapping = append(mapping, []any{"HEAP_ALLOC", fmt.Sprintf("%v MB", memStats.HeapAlloc/1024/1024)})
	mapping = append(mapping, []any{"TOTAL_ALLOC", fmt.Sprintf("%v MB", memStats.TotalAlloc/1024/1024)})
	mapping = append(mapping, []any{"CGO_CALL_NUM", runtime.NumCgoCall()})
	mapping = append(mapping, []any{"GOROUTINE_NUM", runtime.NumGoroutine()})

	str := ""
	for _, v := range mapping {
		str += fmt.Sprintf("%v：%v\n", v[0], v[1])
	}
	return str
}

func main() {
	go listenChannelStatus()

	// ===========消息处理器===========
	msgHandler := network.GetInstanceMsgHandler()
	// 添加一个路由
	msgHandler.AddRouter(int32(pb.MsgId_Echo), func() proto.Message { return &pb.EchoRequest{} }, func(con iface.IConnection, message proto.Message) {
		// do something ...
		req := message.(*pb.EchoRequest)
		// fmt.Println(req.GetMessage())
		res := &pb.EchoResponse{
			Message: req.Message,
		}
		con.SendMsg(int32(pb.MsgId_Echo), res)
	})

	// 注册服务
	network.GetInstanceServerManager().RegisterServer(network.GetServerTCP(), network.GetServerWS(), network.GetServerHTTP())

	fmt.Printf(info())
}
