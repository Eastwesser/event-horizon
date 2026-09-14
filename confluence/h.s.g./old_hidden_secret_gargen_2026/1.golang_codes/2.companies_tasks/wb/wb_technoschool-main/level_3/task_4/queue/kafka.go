package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	kafkago "github.com/segmentio/kafka-go"
	"github.com/wb-go/wbf/kafka"
	"github.com/wb-go/wbf/retry"
	"github.com/wb_technoschool/level_3/task_4/models"
)

type KafkaQueue struct {
	producer *kafka.Producer
	consumer *kafka.Consumer
	topic    string
	brokers  []string
}

func NewKafkaQueue(brokers []string, topic, groupID string) (*KafkaQueue, error) {
	if err := createTopicIfNotExists(brokers, topic); err != nil {
		return nil, fmt.Errorf("failed to create topic: %w", err)
	}

	producer := kafka.NewProducer(brokers, topic)
	consumer := kafka.NewConsumer(brokers, topic, groupID)

	return &KafkaQueue{
		producer: producer,
		consumer: consumer,
		topic:    topic,
		brokers:  brokers,
	}, nil
}

func createTopicIfNotExists(brokers []string, topic string) error {
	conn, err := kafkago.Dial("tcp", brokers[0])
	if err != nil {
		return err
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions(topic)
	if err == nil && len(partitions) > 0 {
		return nil
	}

	controller, err := conn.Controller()
	if err != nil {
		return err
	}

	controllerConn, err := kafkago.Dial("tcp", fmt.Sprintf("%s:%d", controller.Host, controller.Port))
	if err != nil {
		return err
	}
	defer controllerConn.Close()

	topicConfig := kafkago.TopicConfig{
		Topic:             topic,
		NumPartitions:     1,
		ReplicationFactor: 1,
	}

	if err := controllerConn.CreateTopics(topicConfig); err != nil {
		partitions, checkErr := conn.ReadPartitions(topic)
		if checkErr == nil && len(partitions) > 0 {
			return nil
		}
		return err
	}

	return nil
}

func (k *KafkaQueue) Publish(ctx context.Context, task *models.ProcessingTask) error {
	data, err := json.Marshal(task)
	if err != nil {
		return err
	}

	strategy := retry.Strategy{
		Attempts: 3,
		Delay:    1 * time.Second,
		Backoff:  2.0,
	}
	return k.producer.SendWithRetry(ctx, strategy, []byte(task.ImageID), data)
}

func (k *KafkaQueue) Consume(ctx context.Context) (<-chan *models.ProcessingTask, error) {
	taskChan := make(chan *models.ProcessingTask, 100)
	msgChan := make(chan kafkago.Message, 100)

	strategy := retry.Strategy{
		Attempts: 5,
		Delay:    1 * time.Second,
		Backoff:  2.0,
	}

	k.consumer.StartConsuming(ctx, msgChan, strategy)

	go func() {
		defer close(taskChan)
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-msgChan:
				if !ok {
					return
				}

				var task models.ProcessingTask
				if err := json.Unmarshal(msg.Value, &task); err != nil {
					continue
				}

				select {
				case taskChan <- &task:
					k.consumer.Commit(ctx, msg)
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return taskChan, nil
}

func (k *KafkaQueue) Close() error {
	if err := k.producer.Close(); err != nil {
		return err
	}
	return k.consumer.Close()
}
