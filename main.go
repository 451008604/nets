package main

import (
	"fmt"
	"github.com/451008604/nets/iface"
	_ "github.com/451008604/nets/module"
	"github.com/451008604/nets/network"
	pb "github.com/451008604/nets/proto/bin"
	"google.golang.org/protobuf/proto"
	"math/rand"
	"net/http"
	"runtime"
	"time"
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

	// ===========广播管理器===========
	// broadcastManager := network.GetInstanceBroadcastManager()
	// broadcastManager.GetGlobalBroadcastGroup()

	go func() {
		time.Sleep(time.Second * 5)
		go network.GetInstanceConnManager().RangeConnections(func(conn iface.IConnection) {
			time.Sleep(time.Second)
			println(fmt.Sprintf("abc -> %v", conn.GetConnId()))
		})

		go network.GetInstanceConnManager().RangeConnections(func(conn iface.IConnection) {
			println(fmt.Sprintf("123 -> %v", 100+conn.GetConnId()))
		})
	}()

	// ===========连接管理器===========
	network.GetInstanceConnManager().SetConnOnOpened(func(conn iface.IConnection) {
		// do something ...
		println("连接建立", conn.GetConnId())
	})
	network.GetInstanceConnManager().SetConnOnClosed(func(conn iface.IConnection) {
		// do something ...

		time.Sleep(time.Second * time.Duration(3+rand.Intn(2)))
		println("连接断开", conn.GetConnId())
	})
	network.GetInstanceConnManager().SetConnOnRateLimiting(func(conn iface.IConnection) {
		// do something ...
		println("触发限流", conn.RemoteAddrStr())
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
	msgHandler.SetFilter(func(conn iface.IConnection, msgData proto.Message) bool {
		// do something ...
		return true
	})

	// 自定义panic捕获
	msgHandler.SetErrCapture(func(conn iface.IConnection, r any) {
		// do something ...
	})

	network.SetCustomServer(&network.CustomServer{})
	// 注册服务
	network.GetInstanceServerManager().RegisterServer(network.GetServerTCP(), network.GetServerWS())

	fmt.Printf(info())
}
