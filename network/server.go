package network

type server struct {
	serverName string // 服务器名称
	ip         string // IP地址
	port       string // 服务端口
	isClose    bool   // 服务是否已关闭
}

func (s *server) GetServerName() string {
	return s.serverName
}

func (s *server) Start() {
	s.isClose = false
}

func (s *server) Stop() {
	s.isClose = true
}

func (s *server) IsClose() bool {
	return s.isClose
}
