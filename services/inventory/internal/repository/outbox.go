package repository

import (
	"context"

	"github.com/Eastwesser/event-horizon/services/inventory/internal/model"
)

// ItemOutboxWriter persists an item and outbox event in one transaction (PostgreSQL).
type ItemOutboxWriter interface {
	CreateItemWithOutbox(ctx context.Context, item *model.Item, eventType string, eventPayload []byte) error
}
