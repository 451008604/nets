package main

import (
	"github.com/451008604/nets/iface"
	"github.com/451008604/nets/network"
	pb "github.com/451008604/nets/proto/bin"
	"google.golang.org/protobuf/proto"
)

func main() {
	// broadcastManager := network.GetInstanceBroadcastManager()
	connManager := network.GetInstanceConnManager()
	connManager.OnConnOpen(func(conn iface.IConnection) {
		fmt.Println(conn.GetConnID())
	})
	connManager.OnConnClose(func(conn iface.IConnection) {
		fmt.Println(conn.GetConnID())
	})

	msgHandler := network.GetInstanceMsgHandler()
	msgHandler.AddRouter(int32(pb.MSgID_PlayerLogin_Req), func() proto.Message { return &pb.PlayerLoginRequest{} }, func(con iface.IConnection, message proto.Message) {})

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

// func listenChannelStatus() {
// 	goroutineNum := 0
// 	for range time.Tick(time.Second * 1) {
// 		if temp := runtime.NumGoroutine(); temp != goroutineNum {
// 			goroutineNum = temp
// 			fmt.Printf("currentNumberOfThreads: %v\tcurrentNumberOfConnections: %v\n", goroutineNum, network.GetInstanceConnManager().Len())
// 		}
// 	}
// }
