package base

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/451008604/socketServerFrame/config"
	"github.com/451008604/socketServerFrame/logs"
	"github.com/451008604/socketServerFrame/network"
	pb "github.com/451008604/socketServerFrame/proto/bin"
	"google.golang.org/protobuf/proto"
)

// CustomConnect 自定义连接
type CustomConnect struct {
	net.Conn
	address   string // 服务地址
	port      string // 服务端口
	bufferLen uint32 // 消息缓冲区长度
	wg        *sync.WaitGroup
}

// 主线程锁
// var wg *sync.WaitGroup

// 主连接
// var conn *CustomConnect

// 尝试重连次数标识
var restartConnectNum = 0

// NewConnection 新建连接
func (c *CustomConnect) NewConnection(address, port string) {
	// 与服务器请求连接
	serverAddress := address + ":" + port
	dial, err := net.Dial("tcp", serverAddress)
	if logs.PrintLogErr(err, fmt.Sprintf("服务器连接失败：%v 第 %v 次尝试重连中...", serverAddress, restartConnectNum)) {
		restartConnectNum++

		// 与服务器连接失败等待2秒重试，期间会阻塞主进程
		time.Sleep(5 * time.Second)
		c.NewConnection(address, port)
		return
	}
	restartConnectNum = 0

	// 关闭旧的连接
	if c.Conn != nil {
		_ = c.Conn.Close()
	}
	// 创建新的连接
	c.Conn = dial
	c.address = address
	c.port = port
	c.bufferLen = config.GetGlobalObject().MaxPackSize

	// 阻塞主进程
	c.wg = &sync.WaitGroup{}
	c.wg.Add(1)
	// 监听服务器返回的消息
	go func(conn *CustomConnect) {
		conn.wg.Done()
		for {
			receiveData := conn.receiveMsg()
			if receiveData == nil {
				return
			}

			resData := &pb.Ping{}
			_ = proto.Unmarshal(receiveData, resData)
			// 服务器返回的消息
			fmt.Printf("与服务器延迟：%v 微秒\n", resData.GetTimeStamp())
		}
	}(c)
	c.wg.Wait()
	c.wg.Add(1)
}

// SetBlocking 阻塞主进程，等待接受消息
func (c *CustomConnect) SetBlocking() {
	c.wg.Wait()
}

// Disconnect 断开连接，结束主进程
func _() {
	// wg.Done()
	// os.Exit(0)
}

// SendMsg 发送消息到服务器
func (c *CustomConnect) SendMsg(msgId pb.MessageID, msgData []byte) {
	if c == nil {
		return
	}

	// 格式化消息
	dp := network.NewDataPack()
	msg := dp.Pack(network.NewMsgPackage(msgId, msgData))
	_, err := c.Write(msg)

	if logs.PrintLogErr(err, "SendMsg err ") {
		// 重新连接服务器
		c.NewConnection(c.address, c.port)
	}
}

// receiveMsg 接收服务器消息
func (c *CustomConnect) receiveMsg() []byte {
	if c == nil {
		return nil
	}

	dp := network.NewDataPack()
	// 获取消息头信息
	headData := make([]byte, dp.GetHeadLen())
	_, err := io.ReadFull(c.Conn, headData)
	if err != nil {
		return nil
	}
	// 获取消息body
	msgData := dp.Unpack(headData)
	if msgData == nil {
		return nil
	}
	if msgData.GetDataLen() > 0 {
		msgData.SetData(make([]byte, msgData.GetDataLen()))
		_, err = io.ReadFull(c.Conn, msgData.GetData())
		if logs.PrintLogErr(err) {
			return nil
		}
	}
	return msgData.GetData()
}
