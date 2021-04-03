// +build load

package service

import (
	"context"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/itimofeev/simple-billing/internal/app/queue"
	"github.com/itimofeev/simple-billing/internal/app/repository"
)

type task struct {
	ctx      context.Context
	srv      *Service
	userID   int64
	toUserID *int64
	amount   int64
	log      *logrus.Logger
}

// during this test we will send tasks to worker pool that will be processed in few workers, each of them in separate go routine
// we create two users
// user1 with balance=count
// user2 with balance=0
// we create count tasks for deposit 1 to user1 balance
// and count tasks for transfer 1 from user1 to user2 balance
// in the end of test we check that both users have count on their balances
// that means that there were no races
func TestLoad(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	count := 1000

	log := &logrus.Logger{
		Out:          os.Stdout,
		Formatter:    new(logrus.TextFormatter),
		Hooks:        make(logrus.LevelHooks),
		Level:        logrus.DebugLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}

	repo := repository.New("postgresql://postgres:password@localhost:5432/postgres?sslmode=disable")
	q, err := queue.New("nats://localhost:4222")
	require.NoError(t, err)
	srv := New(repo, q)
	ctx := context.Background()

	wp := newWorkerPool(4, 0)

	userID1, userID2 := rand.Int63(), rand.Int63()

	require.NoError(t, srv.CreateAccount(ctx, userID1))
	require.NoError(t, srv.CreateAccount(ctx, userID2))

	require.NoError(t, srv.Deposit(ctx, userID1, int64(count)))

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		for i := 0; i < count; i++ {
			// deposit 1 to userID1
			wp.addTask(task{
				ctx:      ctx,
				srv:      srv,
				userID:   userID1,
				toUserID: nil,
				amount:   1,
				log:      log,
			})
		}
		wg.Done()
	}()

	go func() {
		for i := 0; i < count; i++ {
			// transfer 1 from userID1 to userID2
			wp.addTask(task{
				ctx:      ctx,
				srv:      srv,
				userID:   userID1,
				toUserID: &userID2,
				amount:   1,
				log:      log,
			})
		}
		wg.Done()
	}()

	wg.Wait()

	wp.stop()

	balance1, err := srv.GetBalance(ctx, userID1)
	require.NoError(t, err)
	balance2, err := srv.GetBalance(ctx, userID2)
	require.NoError(t, err)

	require.EqualValues(t, count, balance1.Balance)
	require.EqualValues(t, count, balance2.Balance)
}

type workerPool struct {
	nWorkers int
	taskChan chan task
	wg       *sync.WaitGroup
}

func newWorkerPool(nWorkers int, bufferSize int) *workerPool {
	taskChan := make(chan task, bufferSize)
	wg := &sync.WaitGroup{}
	for i := 0; i < nWorkers; i++ {
		w := worker{taskChan: taskChan, number: i}
		go func() {
			wg.Add(1)
			w.run()
			wg.Done()
		}()
	}

	return &workerPool{
		nWorkers: nWorkers,
		taskChan: taskChan,
		wg:       wg,
	}
}

func (p *workerPool) addTask(t task) {
	p.taskChan <- t
}

func (p *workerPool) stop() {
	close(p.taskChan)
	p.wg.Wait()
}

type worker struct {
	number   int
	taskChan <-chan task
}

func (w *worker) run() {
	for task := range w.taskChan {
		w.processTask(task)
	}
}

func (w *worker) processTask(t task) {
	log := t.log.WithField("worker", w.number)
	if t.toUserID != nil { // transfer
		if err := t.srv.Transfer(t.ctx, t.userID, *t.toUserID, t.amount); err != nil {
			panic(err)
		}
		log.Debug("transferred")
		return
	}

	if t.amount > 0 {
		if err := t.srv.Deposit(t.ctx, t.userID, t.amount); err != nil {
			panic(err)
		}
		log.Debug("deposited")
		return
	}

	if err := t.srv.Withdraw(t.ctx, t.userID, -t.amount); err != nil {
		log.Debug("withdraw")
		panic(err)
	}
}
