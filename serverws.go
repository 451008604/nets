package nets

import (
	"context"
	"github.com/gorilla/websocket"
	"net/http"
)

type serverWS struct {
	manager        *ServerManager
	serverName, ip string
	port           int
}

func (m *ServerManager) GetServerWS() IServer {
	conf := m.config.AppConf.ServerWS
	return &serverWS{m, m.config.AppConf.AppName + "_ws", conf.Address, conf.Port}
}
func GetServerWS() IServer                       { return getDefaultServerManager().GetServerWS() }
func (s *serverWS) managerOwner() *ServerManager { return s.manager }
func (s *serverWS) GetServerName() string        { return s.serverName }
func (s *serverWS) handler() http.Handler {
	upgrade := websocket.Upgrader{ReadBufferSize: int(s.manager.config.AppConf.MaxPackSize), WriteBufferSize: 64 * 1024}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if s.manager.IsClose() {
			http.Error(w, http.StatusText(http.StatusServiceUnavailable), http.StatusServiceUnavailable)
			return
		}
		conn, err := upgrade.Upgrade(w, req, nil)
		if err != nil {
			return
		}
		s.manager.connections.TryAdd(NewConnectionWS(s, conn))
	})
	return mux
}
func (s *serverWS) Start() {
	if err := s.Serve(s.manager.ctx); err != nil {
		s.manager.config.Logger.Error("WebSocket server failed", "err", err)
	}
}
func (s *serverWS) Serve(ctx context.Context) error {
	return serveHTTP(ctx, s.manager, s.manager.config.AppConf.ServerWS, s.handler())
}
