package nets

import (
	"context"
	"errors"
	"io"
	"net/http"
	"sync/atomic"
	"time"
)

const (
	defaultHTTPTimeout = 30 * time.Second
	maxHTTPHeaderBytes = 64 << 10
	httpBodyExtraBytes = 1024
)

type serverHTTP struct {
	manager       *ServerManager
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
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		b.tooLarge = true
	}
	return n, err
}
func (m *ServerManager) GetServerHTTP() IServer {
	conf := m.config.AppConf.ServerHTTP
	return &serverHTTP{manager: m, serverName: m.config.AppConf.AppName + "_http", ip: conf.Address, port: conf.Port}
}
func GetServerHTTP() IServer                       { return getDefaultServerManager().GetServerHTTP() }
func (s *serverHTTP) managerOwner() *ServerManager { return s.manager }
func (s *serverHTTP) GetServerName() string        { return s.serverName }
func (s *serverHTTP) acquireRequest() bool {
	if !s.manager.connections.acquireRequest() {
		return false
	}
	s.activeRequest.Add(1)
	return true
}
func (s *serverHTTP) handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}
		if !s.acquireRequest() {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		defer func() { s.activeRequest.Add(-1); s.manager.connections.releaseRequest() }()
		limit := int64(s.manager.config.AppConf.MaxPackSize)*2 + httpBodyExtraBytes
		if req.ContentLength > limit {
			http.Error(w, http.StatusText(http.StatusRequestEntityTooLarge), http.StatusRequestEntityTooLarge)
			return
		}
		body := &limitedHTTPBody{ReadCloser: http.MaxBytesReader(w, req.Body, limit)}
		req.Body = body
		defer req.Body.Close()
		conn := NewConnectionHTTP(s, w, req)
		if !conn.StartReader() {
			status := http.StatusBadRequest
			if body.tooLarge {
				status = http.StatusRequestEntityTooLarge
			}
			http.Error(w, http.StatusText(status), status)
		}
	})
	return mux
}
func (s *serverHTTP) Start() {
	if err := s.Serve(s.manager.ctx); err != nil {
		s.manager.config.Logger.Error("HTTP server failed", "err", err)
	}
}
func (s *serverHTTP) Serve(ctx context.Context) error {
	return serveHTTP(ctx, s.manager, s.manager.config.AppConf.ServerHTTP, s.handler())
}
