package kafka

import (
	"context"
	"log/slog"
	"sync"
)

// NoopProducer logs and discards (default until Sarama is available: go get + -tags kafka).
type NoopProducer struct {
	Topic  string
	Logger *slog.Logger
}

func (p *NoopProducer) Send(ctx context.Context, key, value []byte) error {
	_ = ctx
	log := p.Logger
	if log == nil {
		log = slog.Default()
	}
	log.Info("kafka noop produce", "topic", p.Topic, "key", string(key), "bytes", len(value))
	return nil
}

func (p *NoopProducer) Close() error { return nil }

// NoopConsumer never delivers messages; Consume blocks until ctx done.
type NoopConsumer struct {
	Topics []string
	Logger *slog.Logger
}

func (c *NoopConsumer) Consume(ctx context.Context, handler MessageHandler) error {
	_ = handler
	log := c.Logger
	if log == nil {
		log = slog.Default()
	}
	log.Info("kafka noop consumer idle", "topics", c.Topics)
	<-ctx.Done()
	return ctx.Err()
}

func (c *NoopConsumer) Close() error { return nil }

// MemoryBroker is an in-process pub/sub for unit tests (not Kafka).
type MemoryBroker struct {
	mu   sync.RWMutex
	subs map[string][]chan Message
}

func NewMemoryBroker() *MemoryBroker {
	return &MemoryBroker{subs: make(map[string][]chan Message)}
}

func (b *MemoryBroker) Publish(topic string, key, value []byte) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	msg := Message{Topic: topic, Key: key, Value: value}
	for _, ch := range b.subs[topic] {
		select {
		case ch <- msg:
		default:
		}
	}
}

func (b *MemoryBroker) Subscribe(topic string) <-chan Message {
	ch := make(chan Message, 64)
	b.mu.Lock()
	b.subs[topic] = append(b.subs[topic], ch)
	b.mu.Unlock()
	return ch
}

// MemoryProducer publishes into MemoryBroker.
type MemoryProducer struct {
	Broker *MemoryBroker
	Topic  string
}

func (p *MemoryProducer) Send(ctx context.Context, key, value []byte) error {
	_ = ctx
	p.Broker.Publish(p.Topic, key, value)
	return nil
}

func (p *MemoryProducer) Close() error { return nil }
