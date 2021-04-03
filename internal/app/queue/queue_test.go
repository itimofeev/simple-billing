package queue

import (
	"testing"
	"time"

	"github.com/nats-io/stan.go"
	"github.com/stretchr/testify/require"
)

func TestPublishSubscribeToNats(t *testing.T) {
	// Connect to a server
	sc, err := stan.Connect("test-cluster", "client-worker-test", stan.NatsURL("nats://localhost:4222"))
	require.NoError(t, err)

	messageReceived := make(chan struct{})
	// Simple Async Subscriber
	_, _ = sc.Subscribe("foo", func(m *stan.Msg) {
		close(messageReceived)
	})

	// Simple Synchronous Publisher
	err = sc.Publish("foo", []byte("Hello World")) // does not return until an ack has been received from NATS Streaming
	require.NoError(t, err)

	select {
	case _, hasMore := <-messageReceived:
		require.False(t, hasMore)
	case <-time.NewTimer(time.Second).C:
		require.Fail(t, "timeout waiting message")
	}
}
