package logic

import (
	"github.com/451008604/socketServerFrame/dao/redis"
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
	// 绑定 Player 和 Conn
	conn.SetPlayer(&Player{Conn: conn})
}

// 断开连接时
func OnConnectionClose(conn iface.IConnection) {
	// 缓存玩家数据
	logs.PrintLogErr(redis.Redis.SetPlayerInfo(GetPlayer(conn).Data.GetCommonData().GetUserID(), GetPlayer(conn).Data), "玩家数据缓存失败")

}
