package queue

import (
	"github.com/nats-io/stan.go"
	"github.com/sirupsen/logrus"
)

const operationCompletedSubject = "operation.completed"

type Queue struct {
	sc  stan.Conn
	log *logrus.Logger
}

// nats://localhost:4222
func New(log *logrus.Logger, url string, clientID string) (*Queue, error) {
	// Connect to a server
	sc, err := stan.Connect("test-cluster", clientID, stan.NatsURL(url))
	if err != nil {
		return nil, err
	}

	return &Queue{sc: sc, log: log}, nil
}

func (q *Queue) Close() error {
	return q.sc.Close()
}
