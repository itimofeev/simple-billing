package queue

import (
	"context"
	"encoding/json"

	"github.com/nats-io/stan.go"

	"github.com/itimofeev/simple-billing/internal/app/model"
)

const operationCompletedSubject = "operation.completed"

type Queue struct {
	sc stan.Conn
}

// nats://localhost:4222
func New(url string) (*Queue, error) {
	// Connect to a server
	sc, err := stan.Connect("test-cluster", "client-worker-test", stan.NatsURL(url))
	if err != nil {
		return nil, err
	}

	return &Queue{sc: sc}, nil
}

func (q *Queue) Publish(_ context.Context, event model.Event, handler stan.AckHandler) (string, error) {
	msgData, err := marshalObject(event)
	if err != nil {
		return "", err
	}
	return q.sc.PublishAsync(operationCompletedSubject, msgData, handler)
}

func marshalObject(object interface{}) ([]byte, error) {
	return json.Marshal(object)
}

func (q *Queue) Close() error {
	return q.sc.Close()
}
