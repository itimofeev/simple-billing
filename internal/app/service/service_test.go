package service

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/itimofeev/simple-billing/internal/app/model"
	"github.com/itimofeev/simple-billing/internal/app/repository"
)

type ServiceSuite struct {
	suite.Suite
	ctx  context.Context
	repo *repository.Repository
	srv  *Service
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}

func (s *ServiceSuite) SetupSuite() {
	rand.Seed(time.Now().UnixNano())
	s.repo = repository.New("postgresql://postgres:password@localhost:5432/postgres?sslmode=disable")
	s.srv = New(s.repo)
	s.ctx = context.Background()
}

func (s *ServiceSuite) Test_ErrorOnGetBalance_IfUserNotFound() {
	userID := rand.Int63()
	_, err := s.srv.GetBalance(s.ctx, userID)
	s.Require().ErrorIs(err, model.ErrUserNotFound)
}

func (s *ServiceSuite) Test_GetBalanceOK_IfUserExists() {
	userID := rand.Int63()
	s.Require().NoError(s.srv.CreateAccount(s.ctx, userID))

	balance, err := s.srv.GetBalance(s.ctx, userID)
	s.Require().NoError(err)

	expected := model.Balance{
		UserID:  userID,
		Balance: 0,
	}

	s.Require().Equal(expected, balance)
}
