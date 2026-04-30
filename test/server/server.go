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
)

var (
	tcpPort  = flag.Int("tcp", 17001, "TCP port")
	wsPort   = flag.Int("ws", 17002, "WebSocket port")
	httpPort = flag.Int("http", 17003, "HTTP port")
	kcpPort  = flag.Int("kcp", 17004, "KCP port")
)

var stats struct {
	flagSend       int32
	flagReceive    int32
	flagOpened     int32
	flagClosed     int32
	flagErrCapture int32
}

func main() {
	flag.Parse()

	go func() {
		_ = http.ListenAndServe(":6060", nil)
	}()

	nets.SetCustomServer(&nets.CustomServer{AppConf: &nets.AppConf{
		ConnRWTimeOut: 60, // 分布式压力测试时适当延长超时时间，避免连接建立后还没有通信就被服务端关闭
		ServerTCP:     nets.ServerConf{Port: *tcpPort},
		ServerWS:      nets.ServerConf{Port: *wsPort},
		ServerHTTP:    nets.ServerConf{Port: *httpPort},
		ServerKCP:     nets.ServerConf{Port: *kcpPort},
	}})
	nets.GetInstanceConnManager().SetConnOpened(func(conn nets.IConnection) { atomic.AddInt32(&stats.flagOpened, 1) })
	nets.GetInstanceConnManager().SetConnClosed(func(conn nets.IConnection) { atomic.AddInt32(&stats.flagClosed, 1) })
	nets.GetInstanceMsgHandler().SetFilter(func(conn nets.IConnection, msg nets.IMessage) bool {
		conn.SetProperty("filterKey", "filterValue")
		return true
	})
	nets.GetInstanceMsgHandler().SetErrCapture(func(conn nets.IConnection, panicInfo string) {
		atomic.AddInt32(&stats.flagErrCapture, 1)
	})

	nets.GetInstanceMsgHandler().AddRouter(int32(internal.Test_MsgId_Test_None), func() proto.Message { return &nets.Message{} }, func(conn nets.IConnection, message proto.Message) {
		reader := conn.GetProperty(nets.ConnPropertyHttpReader).(*http.Request)
		writer := conn.GetProperty(nets.ConnPropertyHttpWriter).(http.ResponseWriter)
		if reader == nil || writer == nil {
			return
		}
		msgReq, ok := message.(*nets.Message)
		if !ok || msgReq == nil {
			return
		}
		conn.SendMsg(int32(internal.Test_MsgId_Test_None), msgReq)
		atomic.AddInt32(&stats.flagReceive, 1)
	})
	nets.GetInstanceMsgHandler().AddRouter(int32(internal.Test_MsgId_Test_Echo), func() proto.Message { return &internal.Test_EchoRequest{} }, func(conn nets.IConnection, message proto.Message) {
		req, ok := message.(*internal.Test_EchoRequest)
		if !ok || req == nil {
			return
		}
		res := &internal.Test_EchoResponse{Message: req.Message}
		conn.SendMsg(int32(internal.Test_MsgId_Test_Echo), res)
		atomic.AddInt32(&stats.flagReceive, 1)
	})

	nets.GetInstanceServerManager().RegisterServer(nets.GetServerHTTP(), nets.GetServerKCP(), nets.GetServerTCP(), nets.GetServerWS())
	fmt.Printf("\nShutting down...\n")
}
