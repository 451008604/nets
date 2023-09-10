package modules

import (
	"github.com/451008604/socketServerFrame/dao/redis"
	"github.com/451008604/socketServerFrame/dao/sql"
	"github.com/451008604/socketServerFrame/iface"
)

var Module = &staticModule{}

type staticModule struct {
	serverTCP iface.IServer        // 服务进程模块
	serverWS  iface.IServer        // 服务进程模块
	notify    iface.INotifyManager // 广播管理模块
	sql       *sql.Module          // sql模块
	redis     *redis.Module        // redis模块
}

func (s *staticModule) ServerTCP() iface.IServer {
	return s.serverTCP
}

func (s *staticModule) SetServerTCP(serverTCP iface.IServer) {
	s.serverTCP = serverTCP
}

func (s *staticModule) ServerWS() iface.IServer {
	return s.serverWS
}

func (s *staticModule) SetServerWS(serverWS iface.IServer) {
	s.serverWS = serverWS
}

func (s *staticModule) Notify() iface.INotifyManager {
	return s.notify
}

func (s *staticModule) SetNotify(notify iface.INotifyManager) {
	s.notify = notify
}
