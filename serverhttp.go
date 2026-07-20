package nets

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	defaultHTTPTimeout = 30 * time.Second
	maxHTTPHeaderBytes = 64 << 10
	httpBodyExtraBytes = 1024
)

type serverHTTP struct {
	serverName    string
	ip            string
	port          int
	activeRequest atomic.Int64
}

type limitedHTTPBody struct {
	io.ReadCloser
	tooLarge bool
}

func (b *limitedHTTPBody) Read(p []byte) (int, error) {
	n, err := b.ReadCloser.Read(p)
	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		b.tooLarge = true
	}
	return n, err
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

func getHTTPTimeout() time.Duration {
	timeout := time.Duration(defaultServer.AppConf.ConnRWTimeOut) * time.Second
	if timeout <= 0 {
		return defaultHTTPTimeout
	}
	return timeout
}

// HTTP 使用 JSON 时 bytes 会经过 base64 编码，因此在 MaxPackSize 基础上预留额外空间。
func getHTTPBodyLimit() int64 {
	maxPackSize := defaultServer.AppConf.MaxPackSize
	if maxPackSize == 0 {
		maxPackSize = 4096
	}
	return int64(maxPackSize)*2 + httpBodyExtraBytes
}

func (s *serverHTTP) acquireRequest() bool {
	maxConn := uint64(defaultServer.AppConf.MaxConn)
	for {
		active := s.activeRequest.Load()
		if uint64(GetInstanceConnManager().Len())+uint64(active) >= maxConn {
			return false
		}
		if s.activeRequest.CompareAndSwap(active, active+1) {
			return true
		}
	}
}

func (s *serverHTTP) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		if GetInstanceServerManager().IsClose() {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}

		if !s.acquireRequest() {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		defer s.activeRequest.Add(-1)

		bodyLimit := getHTTPBodyLimit()
		if r.ContentLength > bodyLimit {
			http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
			return
		}
		limitedBody := &limitedHTTPBody{ReadCloser: http.MaxBytesReader(w, r.Body, bodyLimit)}
		r.Body = limitedBody
		defer func() { _ = r.Body.Close() }()

		msgConn := NewConnectionHTTP(s, w, r)
		if !msgConn.StartReader() {
			status := http.StatusBadRequest
			if limitedBody.tooLarge {
				status = http.StatusRequestEntityTooLarge
			}
			http.Error(w, http.StatusText(status), status)
		}
	})
	return mux
}

func (s *serverHTTP) Start() {
	if s.port == 0 {
		return
	}
	getLogger().Info("server starting", "name", s.serverName, "port", s.port)

	timeout := getHTTPTimeout()
	srv := &http.Server{
		Addr:              net.JoinHostPort(s.ip, strconv.Itoa(s.port)),
		Handler:           s.handler(),
		ReadHeaderTimeout: timeout,
		ReadTimeout:       timeout,
		WriteTimeout:      timeout,
		IdleTimeout:       timeout,
		MaxHeaderBytes:    maxHTTPHeaderBytes,
		ErrorLog:          slog.NewLogLogger(getLogger().Handler(), slog.LevelError),
	}
	go func() {
		<-serverCtx.Done()
		ctx, cancel := context.WithTimeout(context.Background(), defaultHTTPTimeout)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			getLogger().Warn("http server graceful shutdown failed", "name", s.serverName, "err", err)
			if closeErr := srv.Close(); closeErr != nil {
				getLogger().Error("http server forced shutdown failed", "name", s.serverName, "err", closeErr)
			}
		}
	}()

	var err error
	if certPath, keyPath := defaultServer.AppConf.ServerHTTP.TLSCertPath, defaultServer.AppConf.ServerHTTP.TLSKeyPath; certPath != "" && keyPath != "" {
		err = srv.ListenAndServeTLS(certPath, keyPath)
	} else {
		err = srv.ListenAndServe()
	}
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		getLogger().Error("http server stopped unexpectedly", "name", s.serverName, "err", err)
	}
}
