package queue

import (
	"context"

	"github.com/nats-io/stan.go"

	"github.com/itimofeev/simple-billing/internal/app/model"
)

const inputCommandSubject = "input.command"

func (q *Queue) SubscribeCommand(ctx context.Context, f func(ctx context.Context, command model.Command) error) error {
	cb := func(m *stan.Msg) {
		command := model.Command{}
		err := unmarshalObject(m.Data, &command)
		if err != nil {
			q.log.WithError(err).Error("error on unmarshalling command")
			return
		}
		ctx := context.Background()
		if err := f(ctx, command); err != nil {
			q.log.WithError(err).Error("error on calling callback")
			return
		}

		if err := m.Ack(); err != nil {
			q.log.WithError(err).Error("error on acking message")
		}
	}

	opts := []stan.SubscriptionOption{
		stan.DurableName("durableCommand"),
		stan.SetManualAckMode(),
	}
	subscription, err := q.sc.QueueSubscribe(inputCommandSubject, "billing-worker", cb, opts...)
	if err != nil {
		return err
	}

	go unsubscribeIfContextClosed(ctx, subscription)
	return nil
}

func unsubscribeIfContextClosed(ctx context.Context, sub stan.Subscription) {
	<-ctx.Done()
	_ = sub.Close()
}

func (q *Queue) SubscribeOperationCompleted(ctx context.Context, f func(ctx context.Context, event model.Event) error) error {
	cb := func(m *stan.Msg) {
		event := model.Event{}
		err := unmarshalObject(m.Data, &event)
		if err != nil {
			q.log.WithError(err).Error("error on unmarshalling event")
			return
		}
		ctx := context.Background()
		if err := f(ctx, event); err != nil {
			q.log.WithError(err).Error("error on calling callback")
			return
		}
	}

	opts := []stan.SubscriptionOption{
		stan.DurableName("durableEvent"),
	}
	subscription, err := q.sc.Subscribe(operationCompletedSubject, cb, opts...)
	if err != nil {
		return err
	}

	go unsubscribeIfContextClosed(ctx, subscription)
	return nil
}
