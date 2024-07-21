package network

import (
	"github.com/451008604/nets/iface"
	"sync"
)

type workerManager struct {
	workerPool ConcurrentMap[Integer, *TaskModel]
}

type TaskModel struct {
	conn iface.IConnection
	task chan iface.ITaskTemplate
}

var instanceWorkerManager iface.IWorkerManager
var instanceWorkerManagerOnce = sync.Once{}

// 协程池管理器
func GetInstanceWorkerManager() iface.IWorkerManager {
	instanceWorkerManagerOnce.Do(func() {
		instanceWorkerManager = &workerManager{
			workerPool: NewConcurrentStringer[Integer, *TaskModel](),
		}
	})
	return instanceWorkerManager
}

func (w *workerManager) BindTaskQueue(conn iface.IConnection) chan iface.ITaskTemplate {
	workerId := conn.GetConnId() % defaultServer.AppConf.WorkerPoolSize
	taskModel, ok := w.workerPool.Get(Integer(workerId))
	if !ok {
		taskModel = &TaskModel{
			task: make(chan iface.ITaskTemplate, defaultServer.AppConf.WorkerTaskMaxLen),
			conn: conn,
		}
		w.workerPool.Set(Integer(workerId), taskModel)
		// 对工作池进行扩容
		go w.startTaskQueue(taskModel)
	} else {
		// 分配给新的连接时丢弃原有未完成的任务
		for {
			if len(taskModel.task) > 0 {
				<-taskModel.task
			} else {
				break
			}
		}
		taskModel.conn = conn
	}

	return taskModel.task
}

func (w *workerManager) startTaskQueue(taskModel *TaskModel) {
	for template := range taskModel.task {
		template.TaskHandler(taskModel.conn)
	}
}
