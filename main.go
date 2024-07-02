package main

import (
	"fmt"
	"github.com/451008604/nets/iface"
	"github.com/451008604/nets/network"
	pb "github.com/451008604/nets/proto/bin"
	"google.golang.org/protobuf/proto"
	"net/http"
	"runtime"
	"sync/atomic"
)

// 服务指标监控
func listenChannelStatus() {
	serveMux := http.NewServeMux()
	server := &http.Server{Addr: ":17000", Handler: serveMux}
	serveMux.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
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
		mapping = append(mapping, []any{"Flag1", network.Flag1})
		mapping = append(mapping, []any{"Flag2", network.Flag2})
		mapping = append(mapping, []any{"Flag3", network.Flag3})
		mapping = append(mapping, []any{"Flag4", network.Flag4})
		mapping = append(mapping, []any{"Flag5", network.Flag5})
		mapping = append(mapping, []any{"Flag6", network.Flag6})

		str := ""
		for _, v := range mapping {
			str += fmt.Sprintf("%v：%v\n", v[0], v[1])
		}
		_, _ = w.Write([]byte(str))
	})
	fmt.Printf("%v\n", server.ListenAndServe())
}

func main() {
	go listenChannelStatus()

	// ===========广播管理器===========
	// broadcastManager := network.GetInstanceBroadcastManager()
	// broadcastManager.GetGlobalBroadcastGroup()

	// ===========连接管理器===========
	connManager := network.GetInstanceConnManager()
	connManager.OnConnOpen(func(conn iface.IConnection) {
		// do something ...
		println("连接建立", conn.GetConnId())
	})
	connManager.OnConnClose(func(conn iface.IConnection) {
		// do something ...
		atomic.AddUint32(&network.Flag6, 1)
		fmt.Printf("%v\t", network.Flag6)
	})

	// ===========消息处理器===========
	msgHandler := network.GetInstanceMsgHandler()
	// 添加一个路由
	msgHandler.AddRouter(int32(pb.MsgId_Echo_Req), func() proto.Message { return &pb.EchoRequest{} }, func(con iface.IConnection, message proto.Message) {
		// do something ...
		req := message.(*pb.EchoRequest)
		// fmt.Println(req.GetMessage())
		res := &pb.EchoResponse{
			Message: req.Message,
		}
		con.SendMsg(int32(pb.MsgId_Echo_Res), res)
	})

	// 自定义消息过滤器
	msgHandler.SetFilter(func(request iface.IRequest, msgData proto.Message) bool {
		// do something ...
		return true
	})

	// 自定义panic捕获
	msgHandler.SetErrCapture(func(request iface.IRequest, r any) {
		// do something ...
	})

	network.SetCustomServer(&network.CustomServer{})
	// 启动TCP服务
	network.GetServerTCP().Listen()
	// 启动WebSocket服务
	network.GetServerWS().Listen()
	// 阻塞主进程
	network.ServerWaitFlag.Wait()
}
