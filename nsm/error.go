package nsm

import "fmt"

type ErrorCode int

const (
	ErrGeneral         ErrorCode = -1
	ErrIncompatibleAPI ErrorCode = -2
	ErrBlacklisted     ErrorCode = -3
	ErrLaunchFailed    ErrorCode = -4
	ErrNoSuchFile      ErrorCode = -5
	ErrNoSessionOpen   ErrorCode = -6
	ErrUnsavedChanges  ErrorCode = -7
	ErrNotNow          ErrorCode = -8
	ErrBadProject      ErrorCode = -9
	ErrCreateFailed    ErrorCode = -10
)

type Error struct {
	Code ErrorCode
	Msg  string
}

func NewError(code ErrorCode, msg string) *Error {
	return &Error{
		Code: code,
		Msg:  msg,
	}
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s (%d)", e.Msg, e.Code)
}
