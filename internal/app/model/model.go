package model

import "errors"

type Balance struct {
	UserID  int64 `pg:"id,pk"`
	Balance int64 `pg:"balance,notnull,use_zero"`
}

var ErrUserNotFound = errors.New("user not found")
var ErrAlreadyExists = errors.New("user already exists")
