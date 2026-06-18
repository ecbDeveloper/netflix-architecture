package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

type Publisher interface {
	Publish(ctx context.Context, payload interface{}) error
}

type rabbitMQPublisher struct {
	ch        *amqp091.Channel
	queueName string
}

func NewRabbitMQPublisher(ch *amqp091.Channel, queueName string) Publisher {
	return &rabbitMQPublisher{
		ch:        ch,
		queueName: queueName,
	}
}

func (p *rabbitMQPublisher) Publish(ctx context.Context, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}

	err = p.ch.PublishWithContext(
		ctx,
		"",
		p.queueName,
		false,
		false,
		amqp091.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message to amqp channel: %w", err)
	}

	return nil
}
