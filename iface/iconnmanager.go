package iface

// 连接管理器
type IConnManager interface {
	// 生成一个新的连接Id
	NewConnId() int
	// 添加连接
	Add(conn IConnection)
	// 删除连接
	Remove(conn IConnection)
	// 根据ConnId获取连接
	Get(connId int) (IConnection, error)
	// 获取当前连接数量
	Len() int
	// 删除并停止所有连接
	ClearConn()
	// 连接创建时的Hook函数
	OnConnOpen(fun func(conn IConnection))
	// 连接断开时的Hook函数
	OnConnClose(fun func(conn IConnection))
	// 监听并处理操作系统信号
	OperatingSystemSignalHandler()
}
