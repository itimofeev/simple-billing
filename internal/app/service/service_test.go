package service

import (
	"context"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"

	"github.com/itimofeev/simple-billing/internal/app/model"
	"github.com/itimofeev/simple-billing/internal/app/queue"
	"github.com/itimofeev/simple-billing/internal/app/repository"
)

type ServiceSuite struct {
	suite.Suite
	ctx    context.Context
	repo   *repository.Repository
	srv    *Service
	userID int64
	queue  *queue.Queue
	log    *logrus.Logger
}

func TestServiceSuite(t *testing.T) {
	suite.Run(t, new(ServiceSuite))
}

func (s *ServiceSuite) SetupSuite() {
	rand.Seed(time.Now().UnixNano())
	s.repo = repository.New("postgresql://postgres:password@localhost:5432/postgres?sslmode=disable")
	s.log = &logrus.Logger{
		Out:          os.Stdout,
		Formatter:    new(logrus.TextFormatter),
		Hooks:        make(logrus.LevelHooks),
		Level:        logrus.DebugLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}
	q, err := queue.New(s.log, "nats://localhost:4222", "client-worker-load-test")
	s.Require().NoError(err)
	s.queue = q

	s.srv = New(s.repo, s.queue)
	s.ctx = context.Background()
}

func (s *ServiceSuite) TearDownSuite() {
	s.Require().NoError(s.queue.Close())
}

func (s *ServiceSuite) SetupTest() {
	s.userID = rand.Int63()
}

func (s *ServiceSuite) Test_ErrorOnGetBalance_IfUserNotFound() {
	_, err := s.srv.GetBalance(s.ctx, s.userID)
	s.Require().ErrorIs(err, model.ErrUserNotFound)
}

func (s *ServiceSuite) Test_GetBalanceOK_IfUserExists() {
	s.Require().NoError(s.srv.CreateAccount(s.ctx, s.userID))

	balance, err := s.srv.GetBalance(s.ctx, s.userID)
	s.Require().NoError(err)

	expected := model.Balance{
		UserID:  s.userID,
		Balance: 0,
	}

	s.Require().Equal(expected, balance)

	s.checkUserEvents(s.userID, model.EventTypeOpen)
}

func (s *ServiceSuite) Test_Deposit() {
	s.Require().NoError(s.srv.CreateAccount(s.ctx, s.userID))

	err := s.srv.Deposit(s.ctx, s.userID, 10)
	s.Require().NoError(err)

	balance, err := s.srv.GetBalance(s.ctx, s.userID)
	s.Require().NoError(err)

	expected := model.Balance{
		UserID:  s.userID,
		Balance: 10,
	}

	s.Require().Equal(expected, balance)

	s.checkUserEvents(s.userID, model.EventTypeOpen, model.EventTypeDeposit)
}

func (s *ServiceSuite) Test_Withdraw() {
	s.Require().NoError(s.srv.CreateAccount(s.ctx, s.userID))

	err := s.srv.Deposit(s.ctx, s.userID, 10)
	s.Require().NoError(err)

	err = s.srv.Withdraw(s.ctx, s.userID, 3)
	s.Require().NoError(err)

	balance, err := s.srv.GetBalance(s.ctx, s.userID)
	s.Require().NoError(err)

	expected := model.Balance{
		UserID:  s.userID,
		Balance: 7,
	}

	s.Require().Equal(expected, balance)

	s.checkUserEvents(s.userID, model.EventTypeOpen, model.EventTypeDeposit, model.EventTypeWithdraw)

}

func (s *ServiceSuite) Test_ErrorOnWithdraw_IfNegativeBalance() {
	s.Require().NoError(s.srv.CreateAccount(s.ctx, s.userID))

	err := s.srv.Withdraw(s.ctx, s.userID, 3)
	s.Require().ErrorIs(err, model.ErrNegativeBalance)

	s.checkUserEvents(s.userID, model.EventTypeOpen)
}

func (s *ServiceSuite) Test_Transfer() {
	s.Require().NoError(s.srv.CreateAccount(s.ctx, s.userID))
	userID2 := rand.Int63()
	s.Require().NoError(s.srv.CreateAccount(s.ctx, userID2))

	s.Require().NoError(s.srv.Deposit(s.ctx, s.userID, 100))

	s.Require().NoError(s.srv.Transfer(s.ctx, s.userID, userID2, 40))

	balance1, err := s.srv.GetBalance(s.ctx, s.userID)
	s.Require().NoError(err)
	s.Require().EqualValues(60, balance1.Balance)

	balance2, err := s.srv.GetBalance(s.ctx, userID2)
	s.Require().NoError(err)
	s.Require().EqualValues(40, balance2.Balance)

	s.checkUserEvents(s.userID, model.EventTypeOpen, model.EventTypeDeposit, model.EventTypeTransfer)
}

func (s *ServiceSuite) checkUserEvents(userID int64, eventTypes ...model.EventType) {
	events, err := s.repo.ListEventsByFromUserID(s.repo.GetDB(s.ctx), userID)
	s.Require().NoError(err)
	s.Require().Len(events, len(eventTypes))
	for i := range events {
		s.Require().Equal(eventTypes[i], events[i].Type)
	}
}
