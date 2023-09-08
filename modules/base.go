package modules

import (
	"github.com/451008604/socketServerFrame/database/redis"
	"github.com/451008604/socketServerFrame/database/sql"
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

func (s *staticModule) Sql() *sql.Module {
	return s.sql
}

func (s *staticModule) SetSql(sql *sql.Module) {
	s.sql = sql
}

func (s *staticModule) Redis() *redis.Module {
	return s.redis
}

func (s *staticModule) SetRedis(redis *redis.Module) {
	s.redis = redis
}
