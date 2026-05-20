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

	// 1. (可选) 自定义配置参数
	nets.SetCustomServer(&nets.CustomServer{AppConf: &nets.AppConf{
		MaxConn:       1000000,
		ConnRWTimeOut: 5, // 分布式压力测试时适当延长超时时间，避免连接建立后还没有通信就被服务端关闭
		ServerTCP:     nets.ServerConf{Port: *tcpPort},
		ServerWS:      nets.ServerConf{Port: *wsPort},
		ServerHTTP:    nets.ServerConf{Port: *httpPort},
		ServerKCP:     nets.ServerConf{Port: *kcpPort},
	}})

	// 2. (可选) 设置连接建立时Hook函数
	nets.GetInstanceConnManager().SetConnOpened(func(conn nets.IConnection) {
		atomic.AddInt32(&stats.flagOpened, 1)
	})
	// 3. (可选) 设置连接断开时Hook函数
	nets.GetInstanceConnManager().SetConnClosed(func(conn nets.IConnection) {
		atomic.AddInt32(&stats.flagClosed, 1)
	})
	// 4. (可选) 设置过滤器。返回 false 则丢弃消息
	nets.GetInstanceMsgHandler().SetFilter(func(conn nets.IConnection, msg nets.IMessage) bool {
		conn.SetProperty("filterKey", "filterValue")
		return true
	})
	// 5. (可选) 设置连接级别错误捕获
	nets.GetInstanceMsgHandler().SetErrCapture(func(conn nets.IConnection, panicInfo string) {
		atomic.AddInt32(&stats.flagErrCapture, 1)
	})

	// 6. 注册消息路由
	nets.GetInstanceMsgHandler().AddRouter(int32(internal.Test_MsgId_Test_Echo), func() proto.Message { return &internal.Test_EchoRequest{} }, func(conn nets.IConnection, message proto.Message) {
		// 获取请求数据
		msgReq, _ := message.(*internal.Test_EchoRequest)
		// 构造响应数据
		msgRes := &internal.Test_EchoResponse{}
		// 业务处理完毕发送响应数据
		defer conn.SendMsg(int32(internal.Test_MsgId_Test_Echo), msgRes)

		// ...处理逻辑，并设置响应数据...
		msgRes.Message = msgReq.GetMessage()
	})

	// 7. 启动服务(阻塞主协程)
	nets.GetInstanceServerManager().RegisterServer(nets.GetServerHTTP(), nets.GetServerKCP(), nets.GetServerTCP(), nets.GetServerWS())
	fmt.Printf("\nShutting down...\n")
}
