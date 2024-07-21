package module

import (
	"fmt"
	"github.com/451008604/nets/iface"
	"github.com/451008604/nets/network"
	pb "github.com/451008604/nets/proto/bin"
	"time"
)

func init() {
	go ticker()
}

func ticker() {
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
		res := &pb.EchoResponse{
			Message: fmt.Sprintf("second -> %v\n", conn.GetConnId()),
		}
		conn.SendMsg(int32(pb.MsgId_Echo_Res), res)
	})
}

func OnNewMinute(t time.Time) {

}

func OnNewHour(t time.Time) {

}

func OnNewDay(t time.Time) {

}

func OnNewWeek(t time.Time) {

}

func OnNewMonth(t time.Time) {

}
