package nets

import (
	"context"
	"github.com/xtaci/kcp-go"
	"net"
	"strconv"
)

type serverKCP struct {
	manager        *ServerManager
	serverName, ip string
	port           int
}

func (m *ServerManager) GetServerKCP() IServer {
	conf := m.config.AppConf.ServerKCP
	return &serverKCP{m, m.config.AppConf.AppName + "_kcp", conf.Address, conf.Port}
}
func GetServerKCP() IServer                       { return getDefaultServerManager().GetServerKCP() }
func (s *serverKCP) managerOwner() *ServerManager { return s.manager }
func (s *serverKCP) GetServerName() string        { return s.serverName }
func (s *serverKCP) Start() {
	if err := s.Serve(s.manager.ctx); err != nil {
		s.manager.config.Logger.Error("KCP server failed", "err", err)
	}
}
func (s *serverKCP) Serve(ctx context.Context) error {
	if s.port == 0 || ctx.Err() != nil {
		return nil
	}
	listener, err := kcp.Listen(net.JoinHostPort(s.ip, strconv.Itoa(s.port)))
	if err != nil {
		return err
	}
	return serveConnections(ctx, listener, func(conn net.Conn) IConnection {
		if session, ok := conn.(*kcp.UDPSession); ok {
			session.SetStreamMode(true)
			session.SetNoDelay(1, 10, 2, 1)
			session.SetWindowSize(256, 256)
		}
		return NewConnectionKCP(s, conn)
	}, s.manager.connections)
}
