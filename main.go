package main

import (
	"fmt"
	"github.com/451008604/nets/iface"
	"github.com/451008604/nets/network"
	pb "github.com/451008604/nets/proto/bin"
	"google.golang.org/protobuf/proto"
	"runtime"
	"time"
)

func main() {
	go listenChannelStatus()

	// ===========广播管理器===========
	broadcastManager := network.GetInstanceBroadcastManager()
	broadcastManager.GetGlobalBroadcastGroup()

	// ===========连接管理器===========
	connManager := network.GetInstanceConnManager()
	connManager.OnConnOpen(func(conn iface.IConnection) {
		// do something ...
		time.Sleep(time.Second)
		connManager.Remove(conn)
	})
	connManager.OnConnClose(func(conn iface.IConnection) {
		// do something ...
	})

	// ===========消息处理器===========
	msgHandler := network.GetInstanceMsgHandler()
	// 添加一个路由
	msgHandler.AddRouter(int32(pb.MsgId_Echo_Req), func() proto.Message { return &pb.EchoRequest{} }, func(con iface.IConnection, message proto.Message) {
		// do something ...
		req := message.(*pb.EchoRequest)
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
	serverTCP := network.NewServerTCP()
	serverTCP.Listen()
	// 启动WebSocket服务
	serverWS := network.NewServerWS()
	serverWS.Listen()
	// 阻塞主进程
	network.ServerWaitFlag.Wait()
}

func listenChannelStatus() {
	goroutineNum := 0
	for range time.Tick(time.Second * 1) {
		if temp := runtime.NumGoroutine(); temp != goroutineNum {
			goroutineNum = temp
			fmt.Printf("currentNumberOfThreads: %v\tcurrentNumberOfConnections: %v\n", goroutineNum, network.GetInstanceConnManager().Len())
		}
	}
}
