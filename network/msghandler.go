package network

import (
	"fmt"
	"github.com/451008604/nets/iface"
	"sync"
)

type msgHandler struct {
	workQueue  ConcurrentMap[Integer, chan iface.IRequest] // 工作池，每个工作队列中存放等待执行的任务
	apis       map[int32]iface.IRouter                     // 存放每个MsgId所对应处理方法的map属性
	filter     iface.IFilter                               // 消息过滤器
	errCapture iface.IErrCapture                           // 错误捕获器
}

var instanceMsgHandler iface.IMsgHandler
var instanceMsgHandlerOnce = sync.Once{}

// 消息处理器
func GetInstanceMsgHandler() iface.IMsgHandler {
	instanceMsgHandlerOnce.Do(func() {
		instanceMsgHandler = &msgHandler{
			apis:      make(map[int32]iface.IRouter),
			workQueue: NewConcurrentStringer[Integer, chan iface.IRequest](),
		}
	})
	return instanceMsgHandler
}

func (m *msgHandler) DoMsgHandler(request iface.IRequest) {
	defer m.msgErrCapture(request)

	router, ok := m.apis[request.GetMsgId()]
	if !ok {
		return
	}

	msgData := router.GetNewMsg()
	if err := request.GetConnection().ByteToProtocol(request.GetData(), msgData); err != nil {
		fmt.Printf("api msgId %v parsing %v error %v\n", request.GetMsgId(), request.GetData(), err)
		return
	}

	// 限流控制
	if request.GetConnection().FlowControl() {
		fmt.Printf("flowControl RemoteAddress: %v, GetMsgId: %v, GetData: %v\n", request.GetConnection().RemoteAddrStr(), request.GetMsgId(), request.GetData())
		return
	}

	// 过滤器校验
	if m.filter != nil && !m.filter(request, msgData) {
		return
	}

	// 对应的逻辑处理方法
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
	workerId := request.GetConnection().GetWorkId()
	workQueue, ok := m.workQueue.Get(Integer(workerId))
	if !ok {
		workQueue = make(chan iface.IRequest, defaultServer.AppConf.WorkerTaskMaxLen)
		m.workQueue.Set(Integer(workerId), workQueue)
		// 对工作池进行扩容
		go m.startOneWorker(workQueue)
	}

	// 将请求推入worker协程
	workQueue <- request
}

func (m *msgHandler) startOneWorker(workQueue chan iface.IRequest) {
	for req := range workQueue {
		m.DoMsgHandler(req)
	}
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
