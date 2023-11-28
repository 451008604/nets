package network

import (
	"fmt"
	"github.com/451008604/nets/config"
	"sync"

	"github.com/451008604/nets/iface"
)

type msgHandler struct {
	workerPoolSize int                     // 工作池的容量
	workQueue      sync.Map                // 工作池，每个工作队列中存放等待执行的任务
	apis           map[int32]iface.IRouter // 存放每个MsgId所对应处理方法的map属性
	filter         iface.IFilter           // 消息过滤器
	errCapture     iface.IErrCapture       // 错误捕获器
}

var instanceMsgHandler iface.IMsgHandler
var instanceMsgHandlerOnce = sync.Once{}

func GetInstanceMsgHandler() iface.IMsgHandler {
	instanceMsgHandlerOnce.Do(func() {
		instanceMsgHandler = &msgHandler{
			workerPoolSize: config.GetServerConf().WorkerPoolSize,
			apis:           make(map[int32]iface.IRouter),
			workQueue:      sync.Map{},
		}
	})
	return instanceMsgHandler
}

func (m *msgHandler) DoMsgHandler(request iface.IRequest) {
	defer m.msgErrCapture(request)

	router, ok := m.apis[request.GetMsgID()]
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
	if m.filter != nil && !m.filter(request, msgData) {
		return
	}

	router.RunHandler(request, msgData)
}

func (m *msgHandler) AddRouter(msgId int32, msg iface.INewMsgStructTemplate, handler iface.IReceiveMsgHandler) {
	if _, ok := m.apis[msgId]; ok {
		fmt.Printf("msgId is duplicate %v\n", msgId)
	}
	m.apis[msgId] = &baseRouter{}
	m.apis[msgId].SetMsg(msg)
	m.apis[msgId].SetHandler(handler)
}

func (m *msgHandler) GetApis() map[int32]iface.IRouter {
	return m.apis
}

func (m *msgHandler) SendMsgToTaskQueue(request iface.IRequest) {
	// 根据connID平均分配至对应worker
	workerID := request.GetConnection().GetConnID() % m.workerPoolSize
	freeWorkQueueID := m.checkFreeWorkQueue()
	if _, ok := m.workQueue.Load(workerID); !ok && freeWorkQueueID != 0 {
		workerID = freeWorkQueueID
	}
	// 对工作池进行扩容
	workQueue, loaded := m.workQueue.LoadOrStore(workerID, make(chan iface.IRequest, config.GetServerConf().WorkerTaskMaxLen))
	if !loaded {
		go m.startOneWorker(workQueue.(chan iface.IRequest))
	}

	// 将请求推入worker协程
	workQueue.(chan iface.IRequest) <- request
}

func (m *msgHandler) startOneWorker(workQueue chan iface.IRequest) {
	for request := range workQueue {
		m.DoMsgHandler(request)
	}
}

func (m *msgHandler) checkFreeWorkQueue() int {
	freeWorkID := 0
	m.workQueue.Range(func(key, value any) bool {
		if len(value.(chan iface.IRequest)) == 0 {
			freeWorkID = key.(int)
			return false
		}
		return true
	})
	return freeWorkID
}

func (m *msgHandler) SetFilter(fun iface.IFilter) {
	m.filter = fun
}

func (m *msgHandler) SetErrCapture(fun iface.IErrCapture) {
	m.errCapture = fun
}

func (m *msgHandler) msgErrCapture(request iface.IRequest) {
	if m.errCapture == nil {
		return
	}
	if r := recover(); r != nil {
		m.errCapture(request, r)
	}
}
