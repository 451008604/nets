package network

import (
	"errors"
	"fmt"
	"github.com/451008604/nets/config"
	"sync"

	"github.com/451008604/nets/iface"
	"github.com/451008604/nets/logs"
	pb "github.com/451008604/nets/proto/bin"
)

type MsgHandler struct {
	WorkerPoolSize int                        // 工作池的容量
	WorkQueue      sync.Map                   // 工作池，每个工作队列中存放等待执行的任务
	Apis           map[pb.MSgID]iface.IRouter // 存放每个MsgId所对应处理方法的map属性
	Filter         iface.IFilter              // 消息过滤器
	ErrCapture     iface.IErrCapture          // 错误捕获器
}

var instanceMsgHandler iface.IMsgHandler
var instanceMsgHandlerOnce = sync.Once{}

func GetInstanceMsgHandler() iface.IMsgHandler {
	instanceMsgHandlerOnce.Do(func() {
		instanceMsgHandler = &MsgHandler{
			WorkerPoolSize: config.GetGlobalObject().WorkerPoolSize,
			Apis:           make(map[pb.MSgID]iface.IRouter),
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
		logs.PrintLogErr(errors.New(fmt.Sprintf("api msgID %v is not fund", request.GetMsgID())))
		return
	}

	// 对应的逻辑处理方法
	msgData := router.GetNewMsg()
	err := request.GetConnection().ByteToProtocol(request.GetData(), msgData)
	if logs.PrintLogErr(err, fmt.Sprintf("api msgID %v parsing msgData:%v", request.GetMsgID(), request.GetData())) {
		return
	}

	// 过滤器校验
	if m.Filter != nil && !m.Filter(request, msgData) {
		return
	}

	router.RunHandler(request, msgData)
}

// 添加路由，绑定处理函数
func (m *MsgHandler) AddRouter(msgId pb.MSgID, msg iface.INewMsgStructTemplate, handler iface.IReceiveMsgHandler) {
	if _, ok := m.Apis[msgId]; ok {
		logs.PrintLogPanic(errors.New("消息ID重复绑定Handler"))
	}
	m.Apis[msgId] = &BaseRouter{}
	m.Apis[msgId].SetMsg(msg)
	m.Apis[msgId].SetHandler(handler)
}

func (m *MsgHandler) GetApis() map[pb.MSgID]iface.IRouter {
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
	workQueue, loaded := m.WorkQueue.LoadOrStore(workerID, make(chan iface.IRequest, config.GetGlobalObject().WorkerTaskMaxLen))
	if !loaded {
		go m.StartOneWorker(workQueue.(chan iface.IRequest))
	}

	// 将请求推入worker协程
	workQueue.(chan iface.IRequest) <- request
}

// 启动一个工作协程等待处理接收的请求
func (m *MsgHandler) StartOneWorker(workQueue chan iface.IRequest) {
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
