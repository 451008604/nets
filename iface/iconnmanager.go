package iface

// 连接管理器
type IConnManager interface {
	// 生成一个新的连接ID
	NewConnID() int64
	// 添加连接
	Add(conn IConnection)
	// 删除连接
	Remove(conn IConnection)
	// 根据ConnID获取连接
	Get(connID int) (IConnection, error)
	// 获取当前连接数量
	Len() int
	// 删除并停止所有连接
	ClearConn()
	// 连接创建时的Hook函数
	OnConnOpen(fun func(conn IConnection))
	// 调用连接时的Hook函数
	CallbackOnConnOpen(conn IConnection)
	// 连接断开时的Hook函数
	OnConnClose(fun func(conn IConnection))
	// 调用连接断开时的Hook函数
	CallbackOnConnClose(conn IConnection)
}
