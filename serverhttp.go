package nets

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"
)

type serverHTTP struct {
	serverName string
	ip         string
	port       int
}

var (
	serverHttp     IServer
	serverHttpOnce sync.Once
)

func GetServerHTTP() IServer {
	serverHttpOnce.Do(func() {
		serverHttp = &serverHTTP{
			serverName: defaultServer.AppConf.AppName + "_http",
			ip:         defaultServer.AppConf.ServerHTTP.Address,
			port:       defaultServer.AppConf.ServerHTTP.Port,
		}
	})
	return serverHttp
}

func (s *serverHTTP) GetServerName() string {
	return s.serverName
}

func (s *serverHTTP) Start() {
	if s.port == 0 {
		return
	}
	fmt.Printf("server starting [ %v:%v ]\n", s.serverName, s.port)

	httpServer := http.NewServeMux()

	// Full memstats JSON endpoint
	httpServer.HandleFunc("/debug", func(w http.ResponseWriter, r *http.Request) {
		// 仅在显式请求时触发 GC，避免监控请求造成 STW
		if r.URL.Query().Get("gc") == "1" {
			runtime.GC()
		}

		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		w.Header().Set("Content-Type", "application/json")

		var httpcount, tcpcount, wscount, kcpcount int
		GetInstanceConnManager().RangeConnections(func(conn IConnection) {
			switch conn.(type) {
			case *connectionHTTP:
				httpcount++
			case *connectionTCP:
				tcpcount++
			case *connectionWS:
				wscount++
			case *connectionKCP:
				kcpcount++
			}
		})

		maps := map[string]interface{}{
			"time_stamp":    time.Now().Local().Format("2006-01-02 15:04:05"),
			"pid":           os.Getpid(),
			"alloc":         m.Alloc,
			"alloc_total":   m.TotalAlloc,
			"sys":           m.Sys,
			"num_gc":        m.NumGC,
			"num_goroutine": runtime.NumGoroutine(),
			"connections":   GetInstanceConnManager().Len(),
			"count_http":    httpcount,
			"count_tcp":     tcpcount,
			"count_ws":      wscount,
			"count_kcp":     kcpcount,
			"work_pool":     GetInstanceWorkerPool().Stats(),
		}
		marshal, _ := json.MarshalIndent(maps, "", "    ")
		_, _ = fmt.Fprintf(w, "%s", marshal)
	})

	httpServer.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Invalid Request Method / 请求方式非法
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			_, _ = w.Write([]byte(http.StatusText(http.StatusMethodNotAllowed)))
			_ = r.Body.Close()
			return
		}

		// Service Shutdown Status / 服务关闭状态
		if GetInstanceServerManager().IsClose() {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(http.StatusText(http.StatusServiceUnavailable)))
			_ = r.Body.Close()
			return
		}

		// Close new connections when count exceeds limit / 连接数量超过限制后，关闭新建立的连接
		if GetInstanceConnManager().Len() >= int(defaultServer.AppConf.MaxConn) {
			w.WriteHeader(http.StatusGatewayTimeout)
			_, _ = w.Write([]byte(http.StatusText(http.StatusGatewayTimeout)))
			_ = r.Body.Close()
			return
		}

		// Establish new connection and listen for client messages / 建立新连接并监听客户端请求的消息
		msgConn := NewConnectionHTTP(s, w, r)
		// Short connection service does not need read/write goroutine separation / 短链接服务不需要启动读写分离协程
		msgConn.StartReader()
	})

	srv := &http.Server{Addr: fmt.Sprintf("%s:%v", s.ip, s.port), Handler: httpServer}
	go func() {
		<-serverCtx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	if certPath, keyPath := defaultServer.AppConf.ServerHTTP.TLSCertPath, defaultServer.AppConf.ServerHTTP.TLSKeyPath; certPath != "" && keyPath != "" {
		if err := srv.ListenAndServeTLS(certPath, keyPath); err != nil && err != http.ErrServerClosed {
			fmt.Printf("server error: %v\n", err)
		}
	} else {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("server error: %v\n", err)
		}
	}
}
