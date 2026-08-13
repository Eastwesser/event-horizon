package kafka

import "context"

// Producer sends messages to a single topic.
type Producer interface {
	Send(ctx context.Context, key, value []byte) error
	Close() error
}

// Message is one Kafka record delivered to a handler.
type Message struct {
	Topic     string
	Partition int32
	Offset    int64
	Key       []byte
	Value     []byte
}

// MessageHandler processes a consumed message.
type MessageHandler func(ctx context.Context, msg Message) error

// Consumer runs until ctx is cancelled.
type Consumer interface {
	Consume(ctx context.Context, handler MessageHandler) error
	Close() error
}

// Topics used by Event Horizon week-5 purchase flow (Order/Assembly analogue).
const (
	TopicPurchasePaid       = "purchase.paid"
	TopicPurchaseFulfilled  = "purchase.fulfilled"
	ConsumerGroupFulfillment = "fulfillment-service"
	ConsumerGroupShop       = "shop-service"
	ConsumerGroupNotify     = "notification-service"
)
