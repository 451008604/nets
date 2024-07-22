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
		OnNewSecond(t)

		// 每分钟执行
		if s != 0 {
			continue
		}
		OnNewMinute(t)

		// 每小时执行
		if m != 0 {
			continue
		}
		OnNewHour(t)

		// 每天执行
		if h != 0 {
			continue
		}
		OnNewDay(t)

		// 每周执行(周日=0,周六=6)
		if w == 1 {
			OnNewWeek(t)
		}

		// 每月执行
		if d == 1 {
			OnNewMonth(t)
		}
	}
}

func OnNewSecond(t time.Time) {
	network.GetInstanceConnManager().RangeConnections(func(conn iface.IConnection) {
		conn.PushTaskQueue(&OnTick{tickType: TickSecond, time: t})
	})
}

func OnNewMinute(t time.Time) {
	network.GetInstanceConnManager().RangeConnections(func(conn iface.IConnection) {
		conn.PushTaskQueue(&OnTick{tickType: TickMinute, time: t})
	})
}

func OnNewHour(t time.Time) {
	network.GetInstanceConnManager().RangeConnections(func(conn iface.IConnection) {
		conn.PushTaskQueue(&OnTick{tickType: TickHour, time: t})
	})
}

func OnNewDay(t time.Time) {
	network.GetInstanceConnManager().RangeConnections(func(conn iface.IConnection) {
		conn.PushTaskQueue(&OnTick{tickType: TickDay, time: t})
	})
}

func OnNewWeek(t time.Time) {
	network.GetInstanceConnManager().RangeConnections(func(conn iface.IConnection) {
		conn.PushTaskQueue(&OnTick{tickType: TickWeek, time: t})
	})
}

func OnNewMonth(t time.Time) {
	network.GetInstanceConnManager().RangeConnections(func(conn iface.IConnection) {
		conn.PushTaskQueue(&OnTick{tickType: TickMonth, time: t})
	})
}

const (
	TickSecond = iota
	TickMinute
	TickHour
	TickDay
	TickWeek
	TickMonth
)

type OnTick struct {
	tickType int
	time     time.Time
}

func (o *OnTick) TaskHandler(conn iface.IConnection) {
	switch o.tickType {
	case TickSecond:
	case TickMinute:
	case TickHour:
	case TickDay:
	case TickWeek:
	case TickMonth:
	}
}
