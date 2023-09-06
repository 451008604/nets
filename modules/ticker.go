package modules

import "time"

type tickerModule struct {
	tick        <-chan time.Time
	perSecondCh chan time.Time // 每秒钟
	perMinuteCh chan time.Time // 每分钟
	perHourCh   chan time.Time // 每小时
	perDayCh    chan time.Time // 每天
	perWeekCh   chan time.Time // 每周
	perMonthCh  chan time.Time // 每月
}

var ticker = &tickerModule{}

func StartTicker() {
	ticker.tick = time.Tick(time.Second)
	ticker.perSecondCh = make(chan time.Time)
	ticker.perMinuteCh = make(chan time.Time)
	ticker.perHourCh = make(chan time.Time)
	ticker.perDayCh = make(chan time.Time)
	ticker.perWeekCh = make(chan time.Time)
	ticker.perMonthCh = make(chan time.Time)
	go func(ticker *tickerModule) {
		for t := range ticker.tick {
			// 每整秒钟执行一次
			t = t.Truncate(time.Second)
			s, m, h, d, w := t.Second(), t.Minute(), t.Hour(), t.Day(), t.Weekday()

			// 每秒执行
			ticker.perSecondCh <- t

			// 每分钟执行
			if s != 0 {
				continue
			}
			ticker.perMinuteCh <- t

			// 每小时执行
			if m != 0 {
				continue
			}
			ticker.perHourCh <- t

			// 每天执行
			if h != 0 {
				continue
			}
			ticker.perDayCh <- t

			// 每周执行(周日=0,周六=6)
			if w == 1 {
				ticker.perWeekCh <- t
			}

			// 每月执行
			if d == 1 {
				ticker.perMonthCh <- t
			}
		}
	}(ticker)

	onTicker()
}

func onTicker() {
	for {
		select {
		case date := <-ticker.perSecondCh:
			onSecondTicker(date)
		case date := <-ticker.perMinuteCh:
			onMinuteTicker(date)
		case date := <-ticker.perHourCh:
			onHourTicker(date)
		case date := <-ticker.perDayCh:
			onDayTicker(date)
		case date := <-ticker.perWeekCh:
			onWeekTicker(date)
		case date := <-ticker.perMonthCh:
			onMonthTicker(date)
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
