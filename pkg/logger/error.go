package logger

import "errors"

var (
	ErrLoggerNil   = errors.New("logger is nil")
	ErrInvalidEnv  = errors.New("invalid env param: can only takes values 'dev' or 'prod'")
	ErrInvalidType = errors.New("invalid type of logger")
)
