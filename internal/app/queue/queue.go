package queue

import (
	"github.com/nats-io/stan.go"
)

type Queue struct {
	sc stan.Conn
}

// "nats://localhost:4222"
func New(url string) (*Queue, error) {
	// Connect to a server
	sc, err := stan.Connect("test-cluster", "client-worker-test", stan.NatsURL(url))
	if err != nil {
		return nil, err
	}

	return &Queue{sc: sc}, nil
}

func (q *Queue) Close() error {
	return q.sc.Close()
}
