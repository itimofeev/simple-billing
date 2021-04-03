package service

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/go-pg/pg/v10"
	"github.com/nats-io/stan.go"

	"github.com/itimofeev/simple-billing/internal/app/model"
)

type Repository interface {
	GetBalance(tx pg.DBI, userID int64, withLock bool) (model.Balance, error)

	GetDB(ctx context.Context) pg.DBI
	DoInTX(ctx context.Context, f func(tx pg.DBI) error) error
	CreateAccount(tx pg.DBI, userID int64) error
	UpdateBalance(tx pg.DBI, userID, newBalance int64) error

	AddEvent(tx pg.DBI, event model.Event) (model.Event, error)

	SetMessageID(tx pg.DBI, event model.Event, messageID string) error
	SetMessageSent(tx pg.DBI, messageID string, now time.Time) error
}

type Queue interface {
	Publish(_ context.Context, event model.Event, handler stan.AckHandler) (string, error)
}

type Service struct {
	r Repository
	q Queue
}

func New(r Repository, q Queue) *Service {
	return &Service{r: r, q: q}
}

func (s *Service) CreateAccount(ctx context.Context, userID int64) error {
	var event model.Event
	err := s.r.DoInTX(ctx, func(tx pg.DBI) error {
		_, err := s.r.GetBalance(tx, userID, true)
		if err == nil {
			return model.ErrAlreadyExists
		}
		if !errors.Is(err, model.ErrUserNotFound) {
			return err
		}

		if err := s.r.CreateAccount(tx, userID); err != nil {
			return err
		}

		event = model.Event{
			Type:        model.EventTypeOpen,
			FromUserID:  userID,
			CreatedTime: time.Now(),
			QueueID:     strconv.FormatInt(rand.Int63(), 10),
		}
		event, err = s.r.AddEvent(tx, event)
		return err
	})
	if err != nil {
		return err
	}

	return s.SendEvent(ctx, event)
}

func (s *Service) Deposit(ctx context.Context, userID, amount int64) error {
	if amount < 0 {
		return model.ErrNegativeAmount
	}
	var event model.Event
	err := s.r.DoInTX(ctx, func(tx pg.DBI) error {
		balance, err := s.r.GetBalance(tx, userID, true)
		if err != nil {
			return err
		}

		if err := s.r.UpdateBalance(tx, userID, balance.Balance+amount); err != nil {
			return err
		}

		event = model.Event{
			Type:        model.EventTypeDeposit,
			FromUserID:  userID,
			Amount:      &amount,
			CreatedTime: time.Now(),
			QueueID:     strconv.FormatInt(rand.Int63(), 10),
		}
		event, err = s.r.AddEvent(tx, event)
		return err
	})

	if err != nil {
		return err
	}

	return s.SendEvent(ctx, event)
}

func (s *Service) Withdraw(ctx context.Context, userID, amount int64) error {
	if amount < 0 {
		return model.ErrNegativeAmount
	}
	var event model.Event
	err := s.r.DoInTX(ctx, func(tx pg.DBI) error {
		balance, err := s.r.GetBalance(tx, userID, true)
		if err != nil {
			return err
		}
		if balance.Balance-amount < 0 {
			return model.ErrNegativeBalance
		}

		if err := s.r.UpdateBalance(tx, userID, balance.Balance-amount); err != nil {
			return err
		}

		event := model.Event{
			Type:        model.EventTypeWithdraw,
			FromUserID:  userID,
			Amount:      &amount,
			CreatedTime: time.Now(),
			QueueID:     strconv.FormatInt(rand.Int63(), 10),
		}
		event, err = s.r.AddEvent(tx, event)
		return err
	})

	if err != nil {
		return err
	}

	return s.SendEvent(ctx, event)
}

func (s *Service) GetBalance(ctx context.Context, userID int64) (model.Balance, error) {
	return s.r.GetBalance(s.r.GetDB(ctx), userID, false)
}

func (s *Service) Transfer(ctx context.Context, fromUserID, toUserID, amount int64) error {
	if amount < 0 {
		return model.ErrNegativeAmount
	}
	var event model.Event
	err := s.r.DoInTX(ctx, func(tx pg.DBI) error {
		fromBalance, err := s.r.GetBalance(tx, fromUserID, true)
		if err != nil {
			return err
		}
		if fromBalance.Balance-amount < 0 {
			return model.ErrNegativeBalance
		}

		toBalance, err := s.r.GetBalance(tx, toUserID, true)
		if err != nil {
			return err
		}

		if err := s.r.UpdateBalance(tx, fromUserID, fromBalance.Balance-amount); err != nil {
			return err
		}

		if err := s.r.UpdateBalance(tx, toUserID, toBalance.Balance+amount); err != nil {
			return err
		}

		event = model.Event{
			Type:        model.EventTypeTransfer,
			FromUserID:  fromUserID,
			ToUserID:    &toUserID,
			Amount:      &amount,
			CreatedTime: time.Now(),
			QueueID:     strconv.FormatInt(rand.Int63(), 10),
		}
		event, err = s.r.AddEvent(tx, event)
		return err
	})

	if err != nil {
		return err
	}

	return s.SendEvent(ctx, event)
}

func (s *Service) SendEvent(ctx context.Context, event model.Event) error {
	ackHandler := func(messageID string, err error) {
		if err := s.r.SetMessageSent(s.r.GetDB(ctx), messageID, time.Now()); err != nil {
			fmt.Println(err) // log
		}
	}
	messageID, err := s.q.Publish(ctx, event, ackHandler)
	if err != nil {
		return err
	}

	return s.r.SetMessageID(s.r.GetDB(ctx), event, messageID)
}
