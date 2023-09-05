package logs

// 打印模式
var printMode = true

/*
SetPrintMode 设置打印模式

true：打印到控制台，false：打印到文件
*/
func SetPrintMode(v bool) {
	printMode = v
}

// 打印信息
func PrintLogInfo(msg string) {
	if msg == "" {
		return
	}
	if printMode {
		printLogInfoToConsole(msg)
	} else {
		printLogInfoToFile(msg)
	}
}

// 打印错误
func PrintLogErr(err error, tips ...string) bool {
	if err == nil {
		return false
	}
	if printMode {
		return printLogErrToConsole(err, tips...)
	} else {
		return printLogErrToFile(err, tips...)
	}
}

// 打印Panic
func PrintLogPanic(err error) {
	if err == nil {
		return
	}
	if printMode {
		printLogPanicToConsole(err)
	} else {
		printLogPanicToFile(err)
	}
}
