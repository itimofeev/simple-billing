package queue

import (
	"context"
	"encoding/json"

	"github.com/nats-io/stan.go"

	"github.com/itimofeev/simple-billing/internal/app/model"
)

func (q *Queue) PublishOperationCompleted(_ context.Context, event *model.Event, handler stan.AckHandler) (string, error) {
	msgData, err := marshalObject(event)
	if err != nil {
		return "", err
	}
	return q.sc.PublishAsync(operationCompletedSubject, msgData, handler)
}

func (q *Queue) PublishCommand(_ context.Context, command model.Command) error {
	msgData, err := marshalObject(command)
	if err != nil {
		return err
	}
	return q.sc.Publish(inputCommandSubject, msgData)
}

func marshalObject(object interface{}) ([]byte, error) {
	return json.Marshal(object)
}

func unmarshalObject(data []byte, object interface{}) error {
	return json.Unmarshal(data, object)
}
