package natsx

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type Client struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

func Connect(url string) (*Client, error) {
	nc, err := nats.Connect(url, nats.Timeout(5*time.Second), nats.Name("event-horizon-mcp"))
	if err != nil {
		return nil, err
	}
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, err
	}
	return &Client{nc: nc, js: js}, nil
}

func (c *Client) Close() {
	if c != nil && c.nc != nil {
		_ = c.nc.Drain()
	}
}

type StreamInfo struct {
	Name     string   `json:"name"`
	Subjects []string `json:"subjects"`
	Messages uint64   `json:"messages"`
	Bytes    uint64   `json:"bytes"`
	Consumer int      `json:"consumer_count"`
}

func (c *Client) ListStreams(_ context.Context) ([]StreamInfo, error) {
	if c == nil || c.js == nil {
		return nil, fmt.Errorf("nats not connected")
	}
	var out []StreamInfo
	for name := range c.js.StreamNames() {
		si, err := c.js.StreamInfo(name)
		if err != nil {
			continue
		}
		out = append(out, StreamInfo{
			Name:     si.Config.Name,
			Subjects: si.Config.Subjects,
			Messages: si.State.Msgs,
			Bytes:    si.State.Bytes,
			Consumer: si.State.Consumers,
		})
	}
	return out, nil
}

type ConsumerInfo struct {
	Stream   string `json:"stream"`
	Name     string `json:"name"`
	Durable  string `json:"durable"`
	Filter   string `json:"filter_subject"`
	Pending  int    `json:"num_pending"`
	AckFloor uint64 `json:"ack_floor_seq"`
}

func (c *Client) ListConsumers(_ context.Context, stream string) ([]ConsumerInfo, error) {
	if c == nil || c.js == nil {
		return nil, fmt.Errorf("nats not connected")
	}
	var out []ConsumerInfo
	for name := range c.js.ConsumerNames(stream) {
		ci, err := c.js.ConsumerInfo(stream, name)
		if err != nil {
			continue
		}
		out = append(out, ConsumerInfo{
			Stream:   stream,
			Name:     ci.Name,
			Durable:  ci.Config.Durable,
			Filter:   ci.Config.FilterSubject,
			Pending:  int(ci.NumPending),
			AckFloor: ci.AckFloor.Stream,
		})
	}
	return out, nil
}

func JSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
