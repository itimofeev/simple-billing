package service

import (
	"context"
	"errors"

	"github.com/go-pg/pg/v10"

	"github.com/itimofeev/simple-billing/internal/app/model"
)

type Repository interface {
	GetBalance(tx pg.DBI, userID int64) (model.Balance, error)

	GetDB(ctx context.Context) pg.DBI
	DoInTX(ctx context.Context, f func(tx pg.DBI) error) error
	CreateUser(tx pg.DBI, userID int64) error
}

type Service struct {
	r Repository
}

func New(r Repository) *Service {
	return &Service{r: r}
}

func (s *Service) CreateAccount(ctx context.Context, userID int64) error {
	return s.r.DoInTX(ctx, func(tx pg.DBI) error {
		_, err := s.r.GetBalance(tx, userID)
		if err == nil {
			return model.ErrAlreadyExists
		}
		if !errors.Is(err, model.ErrUserNotFound) {
			return err
		}

		return s.r.CreateUser(tx, userID)
	})
}

func (s *Service) Deposit(ctx context.Context, userID, amount int64) error {
	panic("not implemented yet")
}

func (s *Service) GetBalance(ctx context.Context, userID int64) (model.Balance, error) {
	return s.r.GetBalance(s.r.GetDB(ctx), userID)
}
