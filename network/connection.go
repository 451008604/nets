package network

import (
	"fmt"
	"github.com/451008604/socketServerFrame/config"
	"github.com/451008604/socketServerFrame/iface"
	"github.com/451008604/socketServerFrame/logs"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"io"
	"net"

	"sync"
)

type Connection struct {
	Server       iface.IServer          // 当前Conn所属的Server
	conn         interface{}            // 当前连接对象
	ConnID       int                    // 当前连接的ID（SessionID）
	isClosed     bool                   // 当前连接是否已关闭
	MsgHandler   iface.IMsgHandler      // 消息管理MsgId和对应处理函数的消息管理模块
	ExitBuffChan chan bool              // 通知该连接已经退出的channel
	msgBuffChan  chan []byte            // 用于读、写两个goroutine之间的消息通信
	property     map[string]interface{} // 连接属性
	propertyLock sync.RWMutex           // 连接属性读写锁
}

// 新建连接
func NewConnection(server iface.IServer, conn interface{}, msgHandler iface.IMsgHandler) *Connection {
	c := &Connection{
		Server:       server,
		conn:         conn,
		ConnID:       int(server.GetConnMgr().NewConnID()),
		isClosed:     false,
		MsgHandler:   msgHandler,
		ExitBuffChan: make(chan bool, 1),
		msgBuffChan:  make(chan []byte, config.GetGlobalObject().MaxMsgChanLen),
		property:     make(map[string]interface{}),
		propertyLock: sync.RWMutex{},
	}

	// 将新建的连接添加到所属Server的连接管理器内
	server.GetConnMgr().Add(c)

	// 建立连接成功
	logs.PrintLogInfo(fmt.Sprintf("成功建立新的客户端连接 -> %v connID - %v", c.GetConnection().RemoteAddr().String(), c.GetConnID()))
	return c
}

// 获取当前连接对象
func (c *Connection) GetConnection() *net.TCPConn {
	return c.conn.(*net.TCPConn)
}

// 处理conn接收的客户端数据
func (c *Connection) StartReader() {
	defer c.Stop()

	for {
		// 获取客户端的消息头信息
		headData := make([]byte, c.Server.DataPacket().GetHeadLen())
		if _, err := io.ReadFull(c.GetConnection(), headData); err != nil {
			if err != io.EOF {
				logs.PrintLogErr(err)
			}
			return
		}
		// 通过消息头获取dataLen和Id
		msgData := c.Server.DataPacket().Unpack(headData)
		if msgData == nil {
			return
		}
		// 通过消息头获取消息body
		if msgData.GetDataLen() > 0 {
			msgData.SetData(make([]byte, msgData.GetDataLen()))
			if _, err := io.ReadFull(c.GetConnection(), msgData.GetData()); logs.PrintLogErr(err) {
				return
			}
		}

		// 封装请求数据传入处理函数
		req := &Request{conn: c, msg: msgData}
		if config.GetGlobalObject().WorkerPoolSize > 0 {
			c.MsgHandler.SendMsgToTaskQueue(req)
		} else {
			go c.MsgHandler.DoMsgHandler(req)
		}
	}
}

// 写消息goroutine，用户将数据发送给客户端
func (c *Connection) StartWriter() {
	for {
		select {
		case data, ok := <-c.msgBuffChan: // 向客户端发送数据
			if !ok {
				break
			}
			_, err := c.GetConnection().Write(data)
			if logs.PrintLogErr(err, string(data)) {
				break
			}
		case <-c.ExitBuffChan:
			return
		}
	}
}

// 启动连接
func (c *Connection) Start() {
	// 开启用于读的goroutine
	go c.StartReader()
	// 开启用于写的goroutine
	go c.StartWriter()

	c.Server.GetConnMgr().CallbackOnConnOpen(c)

	// 在收到退出消息时释放进程
	for range c.ExitBuffChan {
		return
	}
}

// 停止连接
func (c *Connection) Stop() {
	if c.isClosed {
		return
	}
	c.isClosed = true
	// 通知关闭该连接的监听
	c.ExitBuffChan <- true

	c.Server.GetConnMgr().CallbackOnConnClose(c)

	// 关闭socket连接
	_ = c.GetConnection().Close()
	// 将连接从连接管理器中删除
	c.Server.GetConnMgr().Remove(c)

	// 关闭该连接管道
	close(c.ExitBuffChan)
	close(c.msgBuffChan)
}

// 获取当前连接ID
func (c *Connection) GetConnID() int {
	return c.ConnID
}

// 获取客户端地址信息
func (c *Connection) RemoteAddr() net.Addr {
	return c.GetConnection().RemoteAddr()
}

// 发送消息给客户端
func (c *Connection) SendMsg(msgId pb.MessageID, data []byte) {
	if c.isClosed {
		logs.PrintLogInfo(fmt.Sprintf("连接已关闭导致消息发送失败 -> msgId:%v\tdata:%v", msgId, string(data)))
		return
	}

	// 将消息数据封包
	msg := c.Server.DataPacket().Pack(NewMsgPackage(msgId, data))
	if msg == nil {
		return
	}
	// 写入传输通道发送给客户端
	c.msgBuffChan <- msg
}

// 设置连接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = value
}

// 获取连接属性
func (c *Connection) GetProperty(key string) interface{} {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if value, ok := c.property[key]; ok {
		return value
	} else {
		return nil
	}
}

// 删除连接属性
func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}
