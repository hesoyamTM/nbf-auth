package config

import "errors"

var (
	ErrPathIsEmpty        = errors.New("config path is empty")
	ErrConfigFileNotExist = errors.New("file does not exist")
)
