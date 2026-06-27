package main

import (
	"flag"
	"fmt"
	"github.com/451008604/nets"
	"github.com/451008604/nets/internal"
	"google.golang.org/protobuf/proto"
	"net/http"
	_ "net/http/pprof"
	"sync/atomic"
	"time"
)

var (
	tcpPort  = flag.Int("tcp", 17001, "TCP port")
	wsPort   = flag.Int("ws", 17002, "WebSocket port")
	httpPort = flag.Int("http", 17003, "HTTP port")
	kcpPort  = flag.Int("kcp", 17004, "KCP port")
)

var stats struct {
	flagOpened     int32
	flagClosed     int32
	flagErrCapture int32
}

func main() {
	flag.Parse()

	go func() {
		_ = http.ListenAndServe(":6060", nil)
	}()

	go func() {
		for t := range time.Tick(time.Second) {
			println(t.Format("15:04:05"), "flagOpened: ", atomic.LoadInt32(&stats.flagOpened), ", flagClosed: ", atomic.LoadInt32(&stats.flagClosed), ", flagErrCapture:", atomic.LoadInt32(&stats.flagErrCapture))
		}
	}()

	// 1. (Optional) Custom configuration parameters / 1. (可选) 自定义配置参数
	serverConf := nets.GetServerConf()
	serverConf.MaxConn = 1000000
	serverConf.ConnRWTimeOut = 5
	serverConf.ServerTCP = nets.ServerConf{Port: *tcpPort}
	serverConf.ServerWS = nets.ServerConf{Port: *wsPort}
	serverConf.ServerHTTP = nets.ServerConf{Port: *httpPort}
	serverConf.ServerKCP = nets.ServerConf{Port: *kcpPort}
	nets.SetCustomServer(&nets.CustomServer{AppConf: serverConf})

	// 2. (Optional) Set connection open hook function / 2. (可选) 设置连接建立时Hook函数
	nets.GetInstanceConnManager().SetConnOpened(func(conn nets.IConnection) {
		atomic.AddInt32(&stats.flagOpened, 1)
	})
	// 3. (Optional) Set connection close hook function / 3. (可选) 设置连接断开时Hook函数
	nets.GetInstanceConnManager().SetConnClosed(func(conn nets.IConnection) {
		// 设置3-5秒随机延迟
		// time.Sleep(time.Second * time.Duration(rand.Intn(3)+3))
		atomic.AddInt32(&stats.flagClosed, 1)
	})
	// 4. (Optional) Set filter. Return false to drop message / 4. (可选) 设置过滤器。返回 false 则丢弃消息
	nets.GetInstanceMsgHandler().SetFilter(func(conn nets.IConnection, msg nets.IMessage) bool {
		conn.SetProperty("filterKey", "filterValue")
		return true
	})
	// 5. (Optional) Set connection-level error capture / 5. (可选) 设置连接级别错误捕获
	nets.GetInstanceMsgHandler().SetErrCapture(func(conn nets.IConnection, r any) {
		atomic.AddInt32(&stats.flagErrCapture, 1)
	})

	// 6. Register message router / 6. 注册消息路由
	nets.GetInstanceMsgHandler().AddRouter(int32(internal.Test_MsgId_Test_Echo), func() proto.Message { return &internal.Test_EchoRequest{} }, func(conn nets.IConnection, message proto.Message) {
		// Get request data / 获取请求数据
		msgReq, _ := message.(*internal.Test_EchoRequest)
		// Construct response data / 构造响应数据
		msgRes := &internal.Test_EchoResponse{}
		// Send response data after business processing / 业务处理完毕发送响应数据
		defer conn.SendMsg(int32(internal.Test_MsgId_Test_Echo), msgRes)

		// ...Processing logic and set response data... / ...处理逻辑，并设置响应数据...
		msgRes.Message = msgReq.GetMessage()
	})

	// 7. Start service (block main goroutine) / 7. 启动服务(阻塞主协程)
	nets.GetInstanceServerManager().RegisterServer(nets.GetServerHTTP(), nets.GetServerKCP(), nets.GetServerTCP(), nets.GetServerWS())
	println("----------------\n", time.Now().Format("15:04:05"), "flagOpened: ", atomic.LoadInt32(&stats.flagOpened), ", flagClosed: ", atomic.LoadInt32(&stats.flagClosed), ", flagErrCapture:", atomic.LoadInt32(&stats.flagErrCapture))
	fmt.Printf("\nShutting down...\n")
}
