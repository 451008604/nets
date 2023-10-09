package common

const (
	ErrSuccess = 0 // 成功

	ErrAccountLengthErr      = iota + 1000 // 账户长度错误
	ErrLoginTypeIllegal                    // 登录类型非法
	ErrRegisterFailed                      // 注册失败
	ErrPlayerInfoFetchFailed               // 获取玩家信息失败
)
