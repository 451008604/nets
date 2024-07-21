package iface

// 工作池管理器
type IWorkerManager interface {
	BindTaskQueue(conn IConnection) chan ITaskTemplate
}

// 任务模版
type ITaskTemplate interface {
	TaskHandler(conn IConnection)
}
