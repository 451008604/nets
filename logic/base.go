package logic

import (
	"github.com/451008604/nets/common"
	"github.com/451008604/nets/iface"
	"time"
)

func init() {
	// 启动计时器
	go common.StartTicker(common.TickerCallback{
		OnSecond: onSecondTicker,
		OnMinute: onMinuteTicker,
		OnHour:   onHourTicker,
		OnDay:    onDayTicker,
		OnWeek:   onWeekTicker,
		OnMonth:  onMonthTicker,
	})
}

// 获取连接对应的玩家
func GetPlayer(conn iface.IConnection) *Player {
	player, _ := conn.GetPlayer().(*Player)
	return player
}

// 建立连接时
func OnConnectionOpen(conn iface.IConnection) {
	// 绑定 Player 和 Conn
	conn.SetPlayer(&Player{Conn: conn})
}

// 断开连接时
func OnConnectionClose(conn iface.IConnection) {
	// 缓存玩家数据
	// logs.PrintLogErr(redis.Redis.SetPlayerInfo(GetPlayer(conn).Data.GetCommonData().GetUserUniID(), GetPlayer(conn).Data), "玩家数据缓存失败")
}

// 每秒钟执行
func onSecondTicker(date time.Time) {

}

// 每分钟执行
func onMinuteTicker(date time.Time) {

}

// 每小时执行
func onHourTicker(date time.Time) {

}

// 每天执行
func onDayTicker(date time.Time) {

}

// 每周执行
func onWeekTicker(date time.Time) {

}

// 每月执行
func onMonthTicker(date time.Time) {

}
