package nets

import (
	"context"
	"net"
	"strconv"
)

type serverTCP struct {
	manager        *ServerManager
	serverName, ip string
	port           int
}

func (m *ServerManager) GetServerTCP() IServer {
	conf := m.config.AppConf.ServerTCP
	return &serverTCP{m, m.config.AppConf.AppName + "_tcp", conf.Address, conf.Port}
}
func GetServerTCP() IServer                       { return getDefaultServerManager().GetServerTCP() }
func (s *serverTCP) managerOwner() *ServerManager { return s.manager }
func (s *serverTCP) GetServerName() string        { return s.serverName }
func (s *serverTCP) Start() {
	if err := s.Serve(s.manager.ctx); err != nil {
		s.manager.config.Logger.Error("TCP server failed", "err", err)
	}
}
func (s *serverTCP) Serve(ctx context.Context) error {
	if s.port == 0 || ctx.Err() != nil {
		return nil
	}
	addr, err := net.ResolveTCPAddr("tcp", net.JoinHostPort(s.ip, strconv.Itoa(s.port)))
	if err != nil {
		return err
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}
	return serveConnections(ctx, listener, func(conn net.Conn) IConnection { return NewConnectionTCP(s, conn.(*net.TCPConn)) }, s.manager.connections)
}
