package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/wb-go/wbf/rabbitmq"
	"github.com/wb-go/wbf/retry"
	"github.com/wb_technoschool/level_3/task_1/models"
)

const (
	QueueName    = "notifications"
	ExchangeName = "notifications_exchange"
)

type Queue interface {
	Publish(ctx context.Context, notification *models.Notification) error
	Consume(ctx context.Context) (<-chan *models.Notification, error)
	Close() error
}

type RabbitMQQueue struct {
	conn     *rabbitmq.Connection
	channel  *rabbitmq.Channel
	publisher *rabbitmq.Publisher
	consumer *rabbitmq.Consumer
}

func NewRabbitMQQueue(url string) (*RabbitMQQueue, error) {
	conn, err := rabbitmq.Connect(url, 3, time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open channel: %w", err)
	}

	exchange := rabbitmq.NewExchange(ExchangeName, "direct")
	exchange.Durable = true
	if err := exchange.BindToChannel(channel); err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	queueManager := rabbitmq.NewQueueManager(channel)
	_, err = queueManager.DeclareQueue(QueueName, rabbitmq.QueueConfig{
		Durable: true,
	})
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	err = channel.QueueBind(QueueName, QueueName, ExchangeName, false, nil)
	if err != nil {
		channel.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to bind queue: %w", err)
	}

	publisher := rabbitmq.NewPublisher(channel, ExchangeName)
	consumerConfig := rabbitmq.NewConsumerConfig(QueueName)
	consumerConfig.AutoAck = false
	consumer := rabbitmq.NewConsumer(channel, consumerConfig)

	return &RabbitMQQueue{
		conn:      conn,
		channel:   channel,
		publisher: publisher,
		consumer: consumer,
	}, nil
}

func (q *RabbitMQQueue) Publish(ctx context.Context, notification *models.Notification) error {
	body, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	strategy := retry.Strategy{Attempts: 3, Delay: time.Second, Backoff: 2}
	return q.publisher.PublishWithRetry(body, QueueName, "application/json", strategy)
}

func (q *RabbitMQQueue) Consume(ctx context.Context) (<-chan *models.Notification, error) {
	msgChan := make(chan []byte, 100)
	notifications := make(chan *models.Notification)

	go func() {
		defer close(notifications)

		strategy := retry.Strategy{Attempts: 5, Delay: time.Second, Backoff: 2}
		err := q.consumer.ConsumeWithRetry(msgChan, strategy)
		if err != nil {
			return
		}
	}()

	go func() {
		defer close(notifications)
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgChan:
				if !ok {
					return
				}

				var notification models.Notification
				if err := json.Unmarshal(msg, &notification); err != nil {
					continue
				}

				if time.Now().Before(notification.ScheduledAt) {
					select {
					case <-time.After(time.Until(notification.ScheduledAt)):
					case <-ctx.Done():
						return
					}
				}

				select {
				case notifications <- &notification:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return notifications, nil
}

func (q *RabbitMQQueue) Close() error {
	if q.channel != nil {
		q.channel.Close()
	}
	if q.conn != nil {
		return q.conn.Close()
	}
	return nil
}
