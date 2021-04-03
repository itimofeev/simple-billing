package model

import "time"

type EventType string

const (
	EventTypeOpen     EventType = "open"
	EventTypeDeposit  EventType = "deposit"
	EventTypeWithdraw EventType = "withdraw"
	EventTypeTransfer EventType = "transfer"
)

type Event struct {
	ID int64 `pg:"id,pk" json:"id"`

	Type EventType `pg:"type" json:"type"`

	FromUserID int64  `pg:"from_user_id,notnull" json:"from_user_id"`
	ToUserID   *int64 `pg:"to_user_id" json:"to_user_id"`

	Amount *int64 `pg:"amount" json:"amount"`

	CreatedTime time.Time `pg:"created_time,notnull" json:"created_time"`

	QueueID       string     `pg:"queue_id,notnull" json:"queue_id"`
	QueueSentTime *time.Time `pg:"queue_sent_time" json:"queue_sent_time"`
}
