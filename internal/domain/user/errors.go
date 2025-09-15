package user

import "errors"

var (
	ErrEmptyName    = errors.New("name is empty")
	ErrEmptySurname = errors.New("surname is empty")
	ErrEmptyPhone   = errors.New("phone is empty")
)
