package modules

const (
	ErrSuccess          = 0
	ErrAccountLengthErr = iota + 1000
	ErrLoginTypeIllegal
	ErrRegisterFailed
	ErrPlayerInfoFetchFailed
)
