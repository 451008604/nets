package modules

import (
	"time"
)

func StartTicker() {
	tick := time.Tick(time.Second)
	for t := range tick {
		// 规格化为整秒
		t = t.Truncate(time.Second)
		s, m, h, d, w := t.Second(), t.Minute(), t.Hour(), t.Day(), t.Weekday()

		// 每秒执行
		onSecondTicker(t)

		// 每分钟执行
		if s != 0 {
			continue
		}
		onMinuteTicker(t)

		// 每小时执行
		if m != 0 {
			continue
		}
		onHourTicker(t)

		// 每天执行
		if h != 0 {
			continue
		}
		onDayTicker(t)

		// 每周执行(周日=0,周六=6)
		if w == 1 {
			onWeekTicker(t)
		}

		// 每月执行
		if d == 1 {
			onMonthTicker(t)
		}
	}
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
