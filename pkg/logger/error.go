package logger

import "errors"

var (
	ErrLoggerNil   = errors.New("Logger is nil")
	ErrInvalidEnv  = errors.New("Invalid env param: can only takes values 'dev' or 'prod'")
	ErrInvalidType = errors.New("Invalid type of logger")
)
