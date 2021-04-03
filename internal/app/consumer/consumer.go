package consumer

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"

	"github.com/itimofeev/simple-billing/internal/app/model"
	"github.com/itimofeev/simple-billing/internal/app/queue"
	"github.com/itimofeev/simple-billing/internal/app/service"
)

type Consumer struct {
	log *logrus.Logger
	srv *service.Service
	q   *queue.Queue
}

func New(log *logrus.Logger, srv *service.Service, q *queue.Queue) *Consumer {
	return &Consumer{
		log: log,
		srv: srv,
		q:   q,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	return c.q.SubscribeCommand(ctx, func(ctx context.Context, command model.Command) error {
		log := c.log.WithField("command", command)
		log.Debug("received command")

		var err error

		switch command.Type {
		case model.CommandTypeOpen:
			err = c.srv.CreateAccount(ctx, command.FromUserID)
		case model.CommandTypeDeposit:
			err = c.srv.Deposit(ctx, command.FromUserID, *command.Amount)
		case model.CommandTypeWithdraw:
			err = c.srv.Withdraw(ctx, command.FromUserID, *command.Amount)
		case model.CommandTypeTransfer:
			err = c.srv.Transfer(ctx, command.FromUserID, *command.ToUserID, *command.Amount)
		default:
			err = errors.New("unknown command")
		}

		if err != nil {
			log.WithError(err).Error("error on handling command")
			return err
		}
		log.Info("command handled")
		return nil
	})
}
