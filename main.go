package main

import (
	"fmt"
	"github.com/451008604/nets/iface"
	"github.com/451008604/nets/network"
	pb "github.com/451008604/nets/proto/bin"
	"google.golang.org/protobuf/proto"
	"net/http"
	"runtime"
)

func main() {
	go listenChannelStatus()

	// // ===========广播管理器===========
	// broadcastManager := network.GetInstanceBroadcastManager()
	// broadcastManager.GetGlobalBroadcastGroup()
	//
	// // ===========连接管理器===========
	// connManager := network.GetInstanceConnManager()
	// connManager.OnConnOpen(func(conn iface.IConnection) {
	// 	// do something ...
	// })
	// connManager.OnConnClose(func(conn iface.IConnection) {
	// 	// do something ...
	// })

	// ===========消息处理器===========
	msgHandler := network.GetInstanceMsgHandler()
	// 添加一个路由
	msgHandler.AddRouter(int32(pb.MsgId_Echo_Req), func() proto.Message { return &pb.EchoRequest{} }, func(con iface.IConnection, message proto.Message) {
		// do something ...
		req := message.(*pb.EchoRequest)
		fmt.Println(req.GetMsgId().Number(), req.GetMessage())
		res := &pb.EchoResponse{
			Message: req.Message,
		}
		con.SendMsg(int32(pb.MsgId_Echo_Res), res)
	})

	// // 自定义消息过滤器
	// msgHandler.SetFilter(func(request iface.IRequest, msgData proto.Message) bool {
	// 	// do something ...
	// 	return true
	// })
	//
	// // 自定义panic捕获
	// msgHandler.SetErrCapture(func(request iface.IRequest, r any) {
	// 	// do something ...
	// })

	network.SetCustomServer(&network.CustomServer{})
	// 启动TCP服务
	serverTCP := network.NewServerTCP()
	serverTCP.Listen()
	// 启动WebSocket服务
	serverWS := network.NewServerWS()
	serverWS.Listen()
	// 阻塞主进程
	network.ServerWaitFlag.Wait()
}

// 服务指标监控
func listenChannelStatus() {
	serveMux := http.NewServeMux()
	server := &http.Server{Addr: ":17000", Handler: serveMux}
	serveMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		mapping := make(map[string]any)
		mapping["协程数量"] = runtime.NumGoroutine()

		str := ""
		for k, v := range mapping {
			str += fmt.Sprintf("%v：\t%v\n", k, v)
		}
		_, _ = w.Write([]byte(str))
	})
	fmt.Printf("%v\n", server.ListenAndServe())
}
