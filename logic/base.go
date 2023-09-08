package logic

import (
	"fmt"
	"github.com/451008604/socketServerFrame/iface"
	"github.com/451008604/socketServerFrame/logs"
	"github.com/451008604/socketServerFrame/modules"
)

func init() {
	// 启动计时器
	go modules.StartTicker()
}

// 获取连接对应的玩家
func GetPlayer(conn iface.IConnection) *Player {
	return conn.GetPlayer().(*Player)
}

// 建立连接时
func OnConnectionOpen(conn iface.IConnection) {
	conn.SetPlayer(&Player{})
}

// 断开连接时
func OnConnectionClose(conn iface.IConnection) {
	logs.PrintLogInfo(fmt.Sprintf("客户端%v下线", conn.RemoteAddrStr()))
}
