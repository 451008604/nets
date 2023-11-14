package network

import (
	"fmt"
	"github.com/451008604/nets/config"
	"sync"

	"github.com/451008604/nets/iface"
)

type MsgHandler struct {
	WorkerPoolSize int                     // 工作池的容量
	WorkQueue      sync.Map                // 工作池，每个工作队列中存放等待执行的任务
	Apis           map[int32]iface.IRouter // 存放每个MsgId所对应处理方法的map属性
	Filter         iface.IFilter           // 消息过滤器
	ErrCapture     iface.IErrCapture       // 错误捕获器
}

var instanceMsgHandler iface.IMsgHandler
var instanceMsgHandlerOnce = sync.Once{}

func GetInstanceMsgHandler() iface.IMsgHandler {
	instanceMsgHandlerOnce.Do(func() {
		instanceMsgHandler = &MsgHandler{
			WorkerPoolSize: config.GetServerConf().WorkerPoolSize,
			Apis:           make(map[int32]iface.IRouter),
			WorkQueue:      sync.Map{},
		}
	})
	return instanceMsgHandler
}

// 执行路由绑定的处理函数
func (m *MsgHandler) DoMsgHandler(request iface.IRequest) {
	defer m.msgErrCapture(request)

	router, ok := m.Apis[request.GetMsgID()]
	if !ok {
		fmt.Printf("api msgID is not found %v\n", request.GetMsgID())
		return
	}

	// 对应的逻辑处理方法
	msgData := router.GetNewMsg()
	err := request.GetConnection().ByteToProtocol(request.GetData(), msgData)
	if err != nil {
		fmt.Printf("api msgID %v parsing %v error %v\n", request.GetMsgID(), request.GetData(), err)
		return
	}

	// 过滤器校验
	if m.Filter != nil && !m.Filter(request, msgData) {
		return
	}

	router.RunHandler(request, msgData)
}

// 添加路由，绑定处理函数
func (m *MsgHandler) AddRouter(msgId int32, msg iface.INewMsgStructTemplate, handler iface.IReceiveMsgHandler) {
	if _, ok := m.Apis[msgId]; ok {
		fmt.Printf("msgId is duplicate %v\n", msgId)
	}
	m.Apis[msgId] = &BaseRouter{}
	m.Apis[msgId].SetMsg(msg)
	m.Apis[msgId].SetHandler(handler)
}

func (m *MsgHandler) GetApis() map[int32]iface.IRouter {
	return m.Apis
}

// 将消息发送到任务队列
func (m *MsgHandler) SendMsgToTaskQueue(request iface.IRequest) {
	// 根据connID平均分配至对应worker
	workerID := request.GetConnection().GetConnID() % m.WorkerPoolSize
	freeWorkQueueID := m.checkFreeWorkQueue()
	if _, ok := m.WorkQueue.Load(workerID); !ok && freeWorkQueueID != 0 {
		workerID = freeWorkQueueID
	}
	// 对工作池进行扩容
	workQueue, loaded := m.WorkQueue.LoadOrStore(workerID, make(chan iface.IRequest, config.GetServerConf().WorkerTaskMaxLen))
	if !loaded {
		go m.startOneWorker(workQueue.(chan iface.IRequest))
	}

	// 将请求推入worker协程
	workQueue.(chan iface.IRequest) <- request
}

// 启动一个工作协程等待处理接收的请求
func (m *MsgHandler) startOneWorker(workQueue chan iface.IRequest) {
	for request := range workQueue {
		m.DoMsgHandler(request)
	}
}

func (m *MsgHandler) checkFreeWorkQueue() int {
	freeWorkID := 0
	m.WorkQueue.Range(func(key, value any) bool {
		if len(value.(chan iface.IRequest)) == 0 {
			freeWorkID = key.(int)
			return false
		}
		return true
	})
	return freeWorkID
}

func (m *MsgHandler) SetFilter(fun iface.IFilter) {
	m.Filter = fun
}

func (m *MsgHandler) SetErrCapture(fun iface.IErrCapture) {
	m.ErrCapture = fun
}

func (m *MsgHandler) msgErrCapture(request iface.IRequest) {
	if m.ErrCapture == nil {
		return
	}
	if r := recover(); r != nil {
		m.ErrCapture(request, r)
	}
}
