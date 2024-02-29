package network

import (
	"sync"
)

type server struct {
	serverName string // 服务器名称
	ip         string // IP地址
	port       string // 服务端口
	isClose    bool   // 服务是否已关闭
}

var ServerWaitFlag = &sync.WaitGroup{}

func (s *server) GetServerName() string {
	return s.serverName
}

func (s *server) Start() {
	ServerWaitFlag.Add(1)
}

func (s *server) Stop() {
	GetInstanceConnManager().ClearConn()
	s.isClose = true
	ServerWaitFlag.Done()
}

func (s *server) Listen() bool {
	s.isClose = false
	return false
}
