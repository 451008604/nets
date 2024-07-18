package network

import (
	"fmt"
	"github.com/451008604/nets/iface"
	"sync"
)

type msgHandler struct {
	workerPool ConcurrentMap[Integer, chan iface.IConnection] // 工作池，每个工作队列中存放等待执行的任务
	apis       map[int32]iface.IRouter                        // 存放每个MsgId所对应处理方法的map属性
	filter     iface.IFilter                                  // 消息过滤器
	errCapture iface.IErrCapture                              // 错误捕获器
}

var instanceMsgHandler iface.IMsgHandler
var instanceMsgHandlerOnce = sync.Once{}

// 消息处理器
func GetInstanceMsgHandler() iface.IMsgHandler {
	instanceMsgHandlerOnce.Do(func() {
		manager := &msgHandler{
			apis:       make(map[int32]iface.IRouter),
			workerPool: NewConcurrentStringer[Integer, chan iface.IConnection](),
		}
		instanceMsgHandler = manager
	})
	return instanceMsgHandler
}

func (m *msgHandler) DoMsgHandler(conn iface.IConnection) {
	defer m.msgErrCapture(conn)

	// 连接关闭时丢弃后续所有操作
	if conn.IsClose() {
		return
	}

	msgQueue := conn.GetMsgQueue()
	msg := msgQueue.Remove(msgQueue.Front()).(iface.IMessage)

	router, ok := m.apis[int32(msg.GetMsgId())]
	if !ok {
		return
	}

	msgData := router.GetNewMsg()
	if err := conn.ByteToProtocol(msg.GetData(), msgData); err != nil {
		fmt.Printf("api msgId %v parsing %v error %v\n", msg.GetMsgId(), msg.GetData(), err)
		return
	}

	// 限流控制
	if conn.FlowControl() {
		fmt.Printf("flowControl RemoteAddress: %v, GetMsgId: %v, GetData: %v\n", conn.RemoteAddrStr(), request.GetMsgId(), request.GetData())
		return
	}

	// 过滤器校验
	if m.filter != nil && !m.filter(conn, msgData) {
		return
	}

	// 对应的逻辑处理方法
	router.RunHandler(conn, msgData)
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

func (m *msgHandler) PushInTaskQueue(task iface.ITaskTemplate) {
	workerId := conn.GetWorkId()
	taskQueue, ok := m.workerPool.Get(Integer(workerId))
	if !ok {
		taskQueue = make(chan iface.IConnection, defaultServer.AppConf.WorkerTaskMaxLen)
		m.workerPool.Set(Integer(workerId), taskQueue)
		// 对工作池进行扩容
		go m.startOneWorker(taskQueue)
	}

	// 推入worker协程
	taskQueue <- conn
}

func (m *msgHandler) startOneWorker(workQueue chan iface.IConnection) {
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

func (m *msgHandler) msgErrCapture(conn iface.IConnection) {
	if m.errCapture == nil {
		return
	}
	if r := recover(); r != nil {
		m.errCapture(conn, r)
	}
}
