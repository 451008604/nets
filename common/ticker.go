package common

import (
	"time"
)

type TickerCallback struct {
	OnSecond func(data time.Time)
	OnMinute func(data time.Time)
	OnHour   func(data time.Time)
	OnDay    func(data time.Time)
	OnWeek   func(data time.Time)
	OnMonth  func(data time.Time)
}

func StartTicker(callback TickerCallback) {
	tick := time.Tick(time.Second)
	for t := range tick {
		// 规格化为整秒
		t = t.Truncate(time.Second)
		s, m, h, d, w := t.Second(), t.Minute(), t.Hour(), t.Day(), t.Weekday()

		// 每秒执行
		callback.OnSecond(t)

		// 每分钟执行
		if s != 0 {
			continue
		}
		callback.OnMinute(t)

		// 每小时执行
		if m != 0 {
			continue
		}
		callback.OnHour(t)

		// 每天执行
		if h != 0 {
			continue
		}
		callback.OnDay(t)

		// 每周执行(周日=0,周六=6)
		if w == 1 {
			callback.OnWeek(t)
		}

		// 每月执行
		if d == 1 {
			callback.OnMonth(t)
		}
	}
}
