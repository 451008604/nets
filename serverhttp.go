package nets

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
)

type serverHTTP struct {
	serverName string
	ip         string
	port       int
}

var serverHttp IServer

func GetServerHTTP() IServer {
	if serverHttp == nil {
		serverHttp = &serverHTTP{
			serverName: defaultServer.AppConf.AppName + "_http",
			ip:         defaultServer.AppConf.ServerHTTP.Address,
			port:       defaultServer.AppConf.ServerHTTP.Port,
		}
	}
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
		runtime.GC()

		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, "{\n  \"timestamp\": \"%v\",\n  \"pid\": \"%v\",\n  \"alloc\": \"%v bytes\",\n  \"total_alloc\": \"%v bytes\",\n  \"sys\": \"%v bytes\",\n  \"num_gc\": %v,\n  \"num_goroutine\": %v,\n  \"connections\": %v\n}", getUTCTime().Local().Format("2006-01-02 15:04:05"), os.Getpid(), m.Alloc, m.TotalAlloc, m.Sys, m.NumGC, runtime.NumGoroutine(), GetInstanceConnManager().Len())
	})

	httpServer.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 服务关闭状态
		if GetInstanceServerManager().IsClose() {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte("server is closed"))
			_ = r.Body.Close()
			return
		}

		// 连接数量超过限制后，关闭新建立的连接
		if GetInstanceConnManager().Len() >= defaultServer.AppConf.MaxConn {
			w.WriteHeader(http.StatusGatewayTimeout)
			_, _ = w.Write([]byte("connection max limit"))
			_ = r.Body.Close()
			return
		}

		// 建立新连接并监听客户端请求的消息
		msgConn := NewConnectionHTTP(s, w, r)
		// 短链接服务不需要启动读写分离协程
		msgConn.StartReader()
	})

	if certPath, keyPath := defaultServer.AppConf.ServerHTTP.TLSCertPath, defaultServer.AppConf.ServerHTTP.TLSKeyPath; certPath != "" && keyPath != "" {
		fmt.Printf("%v\n", http.ListenAndServeTLS(fmt.Sprintf("%s:%v", s.ip, s.port), certPath, keyPath, httpServer))
	} else {
		fmt.Printf("%v\n", http.ListenAndServe(fmt.Sprintf("%s:%v", s.ip, s.port), httpServer))
	}
}
