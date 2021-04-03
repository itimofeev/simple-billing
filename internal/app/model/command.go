package model

type CommandType string

const (
	CommandTypeOpen     CommandType = "open"
	CommandTypeDeposit  CommandType = "deposit"
	CommandTypeWithdraw CommandType = "withdraw"
	CommandTypeTransfer CommandType = "transfer"
)

type Command struct {
	ID         int64       `json:"id"`
	Type       CommandType `json:"type"`
	FromUserID int64       `json:"from_user_id"`
	ToUserID   *int64      `json:"to_user_id"`
	Amount     *int64      `json:"amount"`
}
