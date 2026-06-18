package infra

import (
	"fmt"

	"github.com/ecbDeveloper/netflix-architecture/apps/api/internal/config"
	"github.com/rabbitmq/amqp091-go"
)

func InitializeRabbitMQ(cfg *config.Config) (*amqp091.Connection, error) {
	connStr := fmt.Sprintf("amqp://%s:%s@%s:%s/", cfg.RabbitMQUser, cfg.RabbitMQPass, cfg.RabbitMQHost, cfg.RabbitMQPort)
	conn, err := amqp091.Dial(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}
	return conn, nil
}

func DeclareContentQueue(cfg *config.Config, rabbitMQCh *amqp091.Channel) error {
	_, err := rabbitMQCh.QueueDeclare(
		cfg.ContentQueueName,
		true,
		false,
		false,
		false,
		nil,
	)

	return err
}
