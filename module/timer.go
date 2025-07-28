package module

import (
	"github.com/451008604/nets/iface"
	"github.com/451008604/nets/network"
	"time"
)

func init() {
	go tick()
}

func tick() {
	for t := range time.Tick(time.Second) {
		// 规格化为整秒
		t = t.Truncate(time.Second)
		s, m, h, d, w := t.Second(), t.Minute(), t.Hour(), t.Day(), t.Weekday()

		// 每秒执行
		network.GetInstanceConnManager().RangeConnections(func(conn iface.IConnection) { conn.DoTask(func() { OnNewSecond(conn, t) }) })

		// 每分钟执行
		if s != 0 {
			continue
		}
		network.GetInstanceConnManager().RangeConnections(func(conn iface.IConnection) { conn.DoTask(func() { OnNewMinute(conn, t) }) })

		// 每小时执行
		if m != 0 {
			continue
		}
		network.GetInstanceConnManager().RangeConnections(func(conn iface.IConnection) { conn.DoTask(func() { OnNewHour(conn, t) }) })

		// 每天执行
		if h != 0 {
			continue
		}
		network.GetInstanceConnManager().RangeConnections(func(conn iface.IConnection) { conn.DoTask(func() { OnNewDay(conn, t) }) })

		// 每周执行(周日=0,周六=6)
		if w == 1 {
			network.GetInstanceConnManager().RangeConnections(func(conn iface.IConnection) { conn.DoTask(func() { OnNewWeek(conn, t) }) })
		}

		// 每月执行
		if d == 1 {
			network.GetInstanceConnManager().RangeConnections(func(conn iface.IConnection) { conn.DoTask(func() { OnNewMonth(conn, t) }) })
		}
	}
}

func OnNewSecond(c iface.IConnection, t time.Time) {

}

func OnNewMinute(c iface.IConnection, t time.Time) {

}

func OnNewHour(c iface.IConnection, t time.Time) {

}

func OnNewDay(c iface.IConnection, t time.Time) {

}

func OnNewWeek(c iface.IConnection, t time.Time) {

}

func OnNewMonth(c iface.IConnection, t time.Time) {

}
