package main

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"

	"github.com/itimofeev/simple-billing/internal/app/model"
	"github.com/itimofeev/simple-billing/internal/app/queue"
	"github.com/itimofeev/simple-billing/pkg/shutdown"
)

func main() {
	log := newLogger()

	q, err := queue.New(log, "nats://localhost:4222", "client")
	if err != nil {
		log.WithError(err).Panic("error on initializing queue")
	}
	defer q.Close()

	ctx := context.Background()

	eg, ctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		return q.SubscribeOperationCompleted(ctx, func(ctx context.Context, event model.Event) error {
			log.WithField("event", event).Info("operation completed")
			return nil
		})
	})

	sigHandler := shutdown.TermSignalTrap()
	eg.Go(func() error {
		return sigHandler.Wait(ctx)
	})

	eg.Go(func() error {
		err := q.PublishCommand(ctx, model.Command{
			ID:         1,
			Type:       model.CommandTypeOpen,
			FromUserID: 2,
			ToUserID:   nil,
			Amount:     nil,
		})

		if err != nil {
			log.WithError(err).Error("error on publishing command")
			return err
		}
		log.Info("command published")
		return nil
	})

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
