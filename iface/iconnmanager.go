package iface

// 连接管理器
type IConnManager interface {
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
}
