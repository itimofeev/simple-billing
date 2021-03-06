package main

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/itimofeev/simple-billing/internal/app/consumer"
	"github.com/itimofeev/simple-billing/internal/app/queue"
	"github.com/itimofeev/simple-billing/internal/app/repository"
	"github.com/itimofeev/simple-billing/internal/app/service"
	"github.com/itimofeev/simple-billing/pkg/shutdown"
)

func main() {
	log := newLogger()

	repo := repository.New("postgresql://postgres:password@localhost:5432/postgres?sslmode=disable")
	q, err := queue.New(log, "nats://localhost:4222", "worker")
	if err != nil {
		log.WithError(err).Panic("error on initializing queue")
	}
	defer q.Close()

	srv := service.New(repo, q)
	consume := consumer.New(log, srv, q)

	ctx := context.Background()
	eg, ctx := errgroup.WithContext(ctx)

	sigHandler := shutdown.TermSignalTrap()

	eg.Go(func() error {
		return sigHandler.Wait(ctx)
	})

	eg.Go(func() error {
		return consume.Start(ctx)
	})

	log.Info("worker started")

	err = eg.Wait()
	if err != nil && err != shutdown.ErrTermSig && err != context.Canceled {
		log.WithError(err).Panic("errgroup returned error")
	}

	log.Info(ctx, "graceful shutdown successfully finished")
}

func newLogger() *logrus.Logger {
	return &logrus.Logger{
		Out:          os.Stdout,
		Formatter:    new(logrus.TextFormatter),
		Hooks:        make(logrus.LevelHooks),
		Level:        logrus.DebugLevel,
		ExitFunc:     os.Exit,
		ReportCaller: false,
	}
}
