package model

import (
	"errors"
	"time"
)

type Balance struct {
	UserID  int64 `pg:"id,pk"`
	Balance int64 `pg:"balance,notnull,use_zero"`
}

var ErrUserNotFound = errors.New("user not found")
var ErrAlreadyExists = errors.New("user already exists")
var ErrNegativeAmount = errors.New("negative amount")
var ErrNegativeBalance = errors.New("negative balance")

type EventType string

const (
	EventTypeOpen     EventType = "open"
	EventTypeDeposit  EventType = "deposit"
	EventTypeWithdraw EventType = "withdraw"
	EventTypeTransfer EventType = "transfer"
)

type Event struct {
	ID int64 `pg:"id,pk"`

	Type EventType `pg:"type"`

	FromUserID int64  `pg:"from_user_id,notnull"`
	ToUserID   *int64 `pg:"to_user_id"`

	Amount *int64 `pg:"amount"`

	CreatedTime time.Time `pg:"created_time,notnull"`

	QueueID       string     `pg:"queue_id,notnull"`
	QueueSentTime *time.Time `pg:"queue_sent_time"`
}
