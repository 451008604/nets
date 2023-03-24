package logs

import (
	"fmt"
)

type logData struct {
	info  string
	stack string // 堆栈
}

type logErrData struct {
	err   error
	tips  []string
	stack string // 堆栈
}

var (
	logInfoCh  = make(chan logData, 1000)
	logErrCh   = make(chan logErrData, 1000)
	logPanicCh = make(chan error, 1000)
)

func init() {
	go func() {
		for {
			select {
			case msg := <-logInfoCh:
				fmt.Println(msg.stack, msg.info)
			case errInfo := <-logErrCh:
				if len(errInfo.tips) > 0 {
					fmt.Println(errInfo.stack, fmt.Sprintf("%v%v", errInfo.tips, errInfo.err.Error()))
				} else {
					fmt.Println(errInfo.stack, fmt.Sprintf("%v", errInfo.err.Error()))
				}
			case panicInfo := <-logPanicCh:
				panic(panicInfo)
			}
		}
	}()
}

// 打印到控制台信息
func printLogInfoToConsole(msg string) {
	if msg == "" {
		return
	}

	logInfoCh <- logData{
		info:  msg,
		stack: getCallerStack(),
	}
}

// 打印到控制台错误
func printLogErrToConsole(err error, tips ...string) bool {
	if err == nil {
		return false
	}

	logErrCh <- logErrData{
		err:   err,
		tips:  tips,
		stack: getCallerStack(),
	}
	return true
}

// 打印到控制台Panic
func printLogPanicToConsole(err error) {
	if err == nil {
		return
	}

	logPanicCh <- err
}
