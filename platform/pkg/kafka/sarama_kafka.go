//go:build kafka

package kafka

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

func newSaramaProducer(cfg Config, topic string, log *slog.Logger) Producer {
	sc := sarama.NewConfig()
	sc.Producer.Return.Successes = true
	sc.Producer.RequiredAcks = sarama.WaitForAll
	sc.Version = sarama.V3_6_0_0
	sp, err := sarama.NewSyncProducer(cfg.Brokers, sc)
	if err != nil {
		if log != nil {
			log.Error("kafka producer init failed; falling back to noop", "err", err)
		}
		return &NoopProducer{Topic: topic, Logger: log}
	}
	return &saramaProducer{p: sp, topic: topic, log: log}
}

type saramaProducer struct {
	p     sarama.SyncProducer
	topic string
	log   *slog.Logger
}

func (p *saramaProducer) Send(ctx context.Context, key, value []byte) error {
	_ = ctx
	_, _, err := p.p.SendMessage(&sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.ByteEncoder(key),
		Value: sarama.ByteEncoder(value),
	})
	if err != nil && p.log != nil {
		p.log.Error("kafka produce failed", "err", err, "topic", p.topic)
	}
	return err
}

func (p *saramaProducer) Close() error { return p.p.Close() }

func newSaramaConsumer(cfg Config, group string, topics []string, log *slog.Logger) Consumer {
	sc := sarama.NewConfig()
	sc.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}
	sc.Consumer.Offsets.Initial = sarama.OffsetNewest
	sc.Version = sarama.V3_6_0_0
	cg, err := sarama.NewConsumerGroup(cfg.Brokers, group, sc)
	if err != nil {
		if log != nil {
			log.Error("kafka consumer init failed; falling back to noop", "err", err)
		}
		return &NoopConsumer{Topics: topics, Logger: log}
	}
	return &saramaConsumer{group: cg, topics: topics, log: log}
}

type saramaConsumer struct {
	group  sarama.ConsumerGroup
	topics []string
	log    *slog.Logger
}

func (c *saramaConsumer) Consume(ctx context.Context, handler MessageHandler) error {
	h := &groupHandler{handler: handler, log: c.log}
	for {
		if err := c.group.Consume(ctx, c.topics, h); err != nil {
			if strings.Contains(err.Error(), "closed") {
				return nil
			}
			return err
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func (c *saramaConsumer) Close() error { return c.group.Close() }

type groupHandler struct {
	handler MessageHandler
	log     *slog.Logger
}

func (h *groupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *groupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *groupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		m := Message{
			Topic:     msg.Topic,
			Partition: msg.Partition,
			Offset:    msg.Offset,
			Key:       msg.Key,
			Value:     msg.Value,
		}
		if err := h.handler(sess.Context(), m); err != nil {
			if h.log != nil {
				h.log.Error("kafka handler error", "err", err)
			}
			continue
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}
