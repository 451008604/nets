package nets

import (
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"sync"
	"time"
)

type serverWS struct {
	serverName string // Server Name / 服务器名称
	ip         string // IP Address / IP地址
	port       int    // Service Port / 服务端口
}

var (
	serverWs     IServer
	serverWsOnce sync.Once
)

func GetServerWS() IServer {
	serverWsOnce.Do(func() {
		serverWs = &serverWS{
			serverName: defaultServer.AppConf.AppName + "_ws",
			ip:         defaultServer.AppConf.ServerWS.Address,
			port:       defaultServer.AppConf.ServerWS.Port,
		}
	})
	return serverWs
}

func (s *serverWS) GetServerName() string {
	return s.serverName
}

func (s *serverWS) Start() {
	if s.port == 0 {
		return
	}
	fmt.Printf("server starting [ %v:%v ]\n", s.serverName, s.port)

	var upgrade = websocket.Upgrader{
		ReadBufferSize:  int(defaultServer.AppConf.MaxPackSize),
		WriteBufferSize: 64 * 1024,
	}
	wsServer := http.NewServeMux()
	wsServer.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrade.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		// Service Shutdown Status / 服务关闭状态
		if GetInstanceServerManager().IsClose() {
			_ = conn.Close()
			return
		}

		// Close new connections when count exceeds limit / 连接数量超过限制后，关闭新建立的连接
		if GetInstanceConnManager().Len() >= int(defaultServer.AppConf.MaxConn) {
			_ = conn.Close()
			return
		}

		// Establish new connection and listen for client messages / 建立新连接并监听客户端请求的消息
		msgConn := NewConnectionWS(s, conn)
		// Add new connection to unified connection manager / 将新建的连接添加到统一的连接管理器内
		GetInstanceConnManager().Add(msgConn)
	})

	srv := &http.Server{Addr: fmt.Sprintf("%s:%v", s.ip, s.port), Handler: wsServer}
	go func() {
		<-serverCtx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	if certPath, keyPath := defaultServer.AppConf.ServerWS.TLSCertPath, defaultServer.AppConf.ServerWS.TLSKeyPath; certPath != "" && keyPath != "" {
		fmt.Printf("server error: %v\n", srv.ListenAndServeTLS(certPath, keyPath))
	} else {
		fmt.Printf("server error: %v\n", srv.ListenAndServe())
	}
}
