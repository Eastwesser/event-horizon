package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"

	"github.com/Eastwesser/event-horizon/contracts/events"
	"github.com/Eastwesser/event-horizon/platform/pkg/kafka"
	"github.com/Eastwesser/event-horizon/platform/pkg/metrics"
	billingPb "github.com/Eastwesser/event-horizon/services/billing/proto"
	paymentPb "github.com/Eastwesser/event-horizon/services/payment/proto"
	"github.com/Eastwesser/event-horizon/services/shop/internal/model"
	"github.com/Eastwesser/event-horizon/services/shop/internal/repository"
)

// CancelResult is returned by CancelPurchase.
type CancelResult struct {
	NewBalance      int32
	RefundedAmount  int32
	AlreadyRefunded bool
}

// ShopService is the interface for the shop service.
type ShopService interface {
	GetItems(ctx context.Context, category, gameID, userID string) ([]repository.Item, error)
	PurchaseItem(ctx context.Context, userID, itemID string) (int32, error)
	CancelPurchase(ctx context.Context, userID, itemID string) (*CancelResult, error)
	GetInventory(ctx context.Context, userID string) ([]repository.Item, error)
	ListPurchasesByItemIDs(ctx context.Context, itemIDs []string, limit, offset int) ([]repository.PurchaseRecord, int64, int64, int64, error)
	SetKafkaProducer(p kafka.Producer)
}

// ShopStore is the PostgreSQL port used by shopService.
type ShopStore interface {
	GetItems(ctx context.Context, category, gameID string) ([]repository.Item, error)
	GetItemByID(ctx context.Context, itemID string) (*repository.Item, error)
	IsItemOwned(ctx context.Context, userID, itemID string) (bool, error)
	ListOwnedItemIDs(ctx context.Context, userID string) ([]string, error)
	PurchaseItemWithStock(ctx context.Context, userID, itemID, purchaseID string, price int, records []repository.OutboxRecord) error
	NextSpendCycle(ctx context.Context, userID, itemID string) (int, error)
	BumpSpendCycle(ctx context.Context, userID, itemID string) error
	SavePurchaseSaga(ctx context.Context, saga repository.PurchaseSaga) (string, string, error)
	MarkPurchaseSaga(ctx context.Context, id, state string) error
	EnqueueOutbox(ctx context.Context, eventType string, payload []byte) (string, error)
	MarkOutboxProcessed(ctx context.Context, id string) error
	GetUserInventory(ctx context.Context, userID string) ([]repository.Item, error)
	RefundPurchase(ctx context.Context, userID, itemID string) (*repository.RefundResult, error)
	MarkPurchaseFulfilled(ctx context.Context, purchaseID string) error
	ListPurchasesByItemIDs(ctx context.Context, itemIDs []string, limit, offset int) ([]repository.PurchaseRecord, int64, int64, int64, error)
	CreateItemFromInventory(ctx context.Context, itemID, name, description string, price float64, stock int) error
}

// ShopCache is the Redis port for catalog lists / invalidation.
type ShopCache interface {
	GetItems(ctx context.Context, key string) ([]repository.Item, error)
	SetItems(ctx context.Context, key string, items []repository.Item, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

// Note: kafka is replaced with NATS in our project
type shopService struct {
	pgRepo    ShopStore
	redisRepo ShopCache
	js        nats.JetStreamContext
	billing   billingPb.BillingServiceClient
	payment   paymentPb.PaymentServiceClient
	kafkaProd kafka.Producer // optional (noop if unset / KAFKA_BROKERS empty)
}

// New constructs the shop service. gRPC dial / NATS inventory sync live in app wiring.
func New(
	pg ShopStore,
	redis ShopCache,
	js nats.JetStreamContext,
	billing billingPb.BillingServiceClient,
	payment paymentPb.PaymentServiceClient,
) ShopService {
	return &shopService{
		pgRepo:    pg,
		redisRepo: redis,
		js:        js,
		billing:   billing,
		payment:   payment,
	}
}

func (s *shopService) GetItems(ctx context.Context, category, gameID, userID string) ([]repository.Item, error) {
	cacheKey := fmt.Sprintf("shop:items:%s:%s", category, gameID)

	// Пытаемся получить из Redis
	items, err := s.redisRepo.GetItems(ctx, cacheKey)
	if err == nil {
		s.applyOwnedFlags(ctx, userID, items)
		return items, nil
	}

	// Если нет в кеше — из PostgreSQL
	items, err = s.pgRepo.GetItems(ctx, category, gameID)
	if err != nil {
		return nil, err
	}

	// Сохраняем в Redis (TTL 5 минут)
	_ = s.redisRepo.SetItems(ctx, cacheKey, items, 5*time.Minute)

	s.applyOwnedFlags(ctx, userID, items)
	return items, nil
}

// applyOwnedFlags sets Owned from one inventory query (avoids N+1 IsItemOwned).
func (s *shopService) applyOwnedFlags(ctx context.Context, userID string, items []repository.Item) {
	if len(items) == 0 || userID == "" {
		return
	}
	ids, err := s.pgRepo.ListOwnedItemIDs(ctx, userID)
	if err != nil {
		return
	}
	owned := make(map[string]struct{}, len(ids))
	for _, id := range ids {
		owned[id] = struct{}{}
	}
	for i := range items {
		_, items[i].Owned = owned[items[i].ID]
	}
}

// PurchaseItem spends tickets in billing, then writes stock and the outbox in the shop database.
// The returned int32 is the ticket balance after a successful spend.
func (s *shopService) PurchaseItem(ctx context.Context, userID, itemID string) (int32, error) {
	item, err := s.pgRepo.GetItemByID(ctx, itemID)
	if err != nil {
		return 0, model.ErrItemNotFound
	}
	if !item.Available {
		return 0, model.ErrItemUnavailable
	}

	if item.Category == "merch" {
		if err := s.checkMerchAllowed(ctx, userID); err != nil {
			return 0, err
		}
	}

	owned, err := s.pgRepo.IsItemOwned(ctx, userID, itemID)
	if err != nil {
		return 0, err
	}
	if owned {
		return 0, model.ErrAlreadyOwned
	}

	cycle, err := s.pgRepo.NextSpendCycle(ctx, userID, itemID)
	if err != nil {
		return 0, fmt.Errorf("spend cycle: %w", err)
	}
	spendRef := fmt.Sprintf("shop:%s:%s:%d", userID, itemID, cycle)
	refundRef := fmt.Sprintf("refund:shop:%s:%s:%d", userID, itemID, cycle)
	sagaID, purchaseID, err := s.pgRepo.SavePurchaseSaga(ctx, repository.PurchaseSaga{
		ID: uuid.NewString(), UserID: userID, ItemID: itemID, PurchaseID: uuid.NewString(),
		SpendRef: spendRef, RefundRef: refundRef, Price: item.Price, State: "started",
	})
	if err != nil {
		return 0, fmt.Errorf("purchase saga: %w", err)
	}

	spendResp, err := s.billing.SpendCurrency(ctx, &billingPb.SpendCurrencyRequest{
		UserId:      userID,
		Currency:    billingPb.CurrencyType_TICKETS,
		Amount:      int32(item.Price),
		Reason:      "shop_purchase",
		ReferenceId: spendRef,
		CheckOnly:   false,
	})
	alreadySpent := false
	if err != nil {
		if strings.Contains(err.Error(), "insufficient balance") {
			_ = s.pgRepo.MarkPurchaseSaga(ctx, sagaID, "aborted")
			return 0, model.ErrInsufficientFunds
		}
		if !isDuplicateRef(err) {
			_ = s.pgRepo.MarkPurchaseSaga(ctx, sagaID, "aborted")
			return 0, fmt.Errorf("failed to spend tickets: %w", err)
		}
		alreadySpent = true
	} else if spendResp == nil || !spendResp.GetSuccess() {
		msg := "spend rejected"
		if spendResp != nil && spendResp.GetMessage() != "" {
			msg = spendResp.GetMessage()
		}
		if strings.Contains(msg, "insufficient balance") {
			_ = s.pgRepo.MarkPurchaseSaga(ctx, sagaID, "aborted")
			return 0, model.ErrInsufficientFunds
		}
		if !isDuplicateRef(errors.New(msg)) {
			_ = s.pgRepo.MarkPurchaseSaga(ctx, sagaID, "aborted")
			return 0, fmt.Errorf("failed to spend tickets: %s", msg)
		}
		alreadySpent = true
	}
	_ = s.pgRepo.MarkPurchaseSaga(ctx, sagaID, "spent")

	event := map[string]interface{}{
		"user_id":     userID,
		"item_id":     itemID,
		"item_name":   item.Name,
		"price":       item.Price,
		"category":    item.Category,
		"purchase_id": purchaseID,
		"timestamp":   time.Now().Unix(),
	}
	eventData, _ := json.Marshal(event)
	paid := events.PurchasePaid{
		EventUUID:    uuid.NewString(),
		PurchaseUUID: purchaseID,
		UserUUID:     userID,
		ItemUUID:     itemID,
		Price:        int32(item.Price),
	}
	paidBody, mErr := paid.Marshal()
	if mErr != nil {
		return 0, fmt.Errorf("marshal purchase.paid: %w", mErr)
	}

	if err := s.pgRepo.PurchaseItemWithStock(ctx, userID, itemID, purchaseID, item.Price, []repository.OutboxRecord{
		{EventType: "shop.purchased", Payload: eventData},
		{EventType: kafka.TopicPurchasePaid, Payload: paidBody},
	}); err != nil {
		if isDuplicateRef(err) {
			_ = s.pgRepo.MarkPurchaseSaga(ctx, sagaID, "stocked")
			return 0, model.ErrAlreadyOwned
		}
		_ = s.pgRepo.MarkPurchaseSaga(ctx, sagaID, "compensating")
		log.Printf("CRITICAL: tickets spent but purchase failed for user %s, item %s: %v — refunding", userID, itemID, err)
		outboxID := s.enqueueRefund(ctx, userID, itemID, item.Price, refundRef, err)
		_ = s.pgRepo.BumpSpendCycle(ctx, userID, itemID)
		if _, refundErr := s.billing.AddCurrency(ctx, &billingPb.AddCurrencyRequest{
			UserId:      userID,
			Currency:    billingPb.CurrencyType_TICKETS,
			Amount:      int32(item.Price),
			Reason:      "shop_purchase_refund",
			ReferenceId: refundRef,
		}); refundErr != nil && !isDuplicateRef(refundErr) {
			metrics.PurchaseRefundFailures.Inc()
			log.Printf("CRITICAL: refund also failed for user %s, item %s: %v", userID, itemID, refundErr)
			return 0, fmt.Errorf("failed to record purchase: %w", err)
		}
		if outboxID != "" {
			_ = s.pgRepo.MarkOutboxProcessed(ctx, outboxID)
		}
		_ = s.pgRepo.MarkPurchaseSaga(ctx, sagaID, "compensated")
		return 0, fmt.Errorf("failed to record purchase: %w", err)
	}
	_ = s.pgRepo.MarkPurchaseSaga(ctx, sagaID, "stocked")
	if alreadySpent {
		log.Printf("shop: replayed spend %s for user %s item %s", spendRef, userID, itemID)
	}

	gameID := ""
	if item.GameID != nil {
		gameID = *item.GameID
	}
	_ = s.redisRepo.Delete(ctx, fmt.Sprintf("shop:items:%s:%s", item.Category, gameID))
	_ = s.redisRepo.Delete(ctx, fmt.Sprintf("shop:items:%s:", item.Category))
	_ = s.redisRepo.Delete(ctx, "shop:items:all:")
	_ = s.redisRepo.Delete(ctx, fmt.Sprintf("balance:%s:tickets", userID))

	metrics.RecordOrder(float64(item.Price))

	if spendResp != nil {
		return spendResp.NewBalance, nil
	}
	bal, balErr := s.billing.GetBalance(ctx, &billingPb.GetBalanceRequest{
		UserId:   userID,
		Currency: billingPb.CurrencyType_TICKETS,
	})
	if balErr != nil || bal == nil {
		return 0, nil
	}
	return bal.GetBalance(), nil
}

// enqueueRefund stores a billing.refund outbox row so a crashed process can still return the tickets.
// An empty id means the insert failed and only the in-process refund remains.
func (s *shopService) enqueueRefund(ctx context.Context, userID, itemID string, price int, referenceID string, cause error) string {
	body, _ := json.Marshal(map[string]any{
		"user_id":      userID,
		"item_id":      itemID,
		"amount":       price,
		"reference_id": referenceID,
		"error":        cause.Error(),
	})
	id, err := s.pgRepo.EnqueueOutbox(ctx, "billing.refund", body)
	if err != nil {
		log.Printf("shop: enqueue refund %s: %v", referenceID, err)
		return ""
	}
	return id
}

// creditCancel pays tickets back for a cancelled purchase.
// reference_id is cancel-{purchaseID}, so a repeated cancel does not credit twice.
func (s *shopService) creditCancel(ctx context.Context, userID string, refund *repository.RefundResult) (int32, error) {
	if refund == nil || refund.PurchaseID == "" || refund.Price <= 0 {
		bal, err := s.billing.GetBalance(ctx, &billingPb.GetBalanceRequest{
			UserId:   userID,
			Currency: billingPb.CurrencyType_TICKETS,
		})
		if err != nil || bal == nil {
			return 0, err
		}
		return bal.GetBalance(), nil
	}
	refID := fmt.Sprintf("cancel-%s", refund.PurchaseID)
	addResp, err := s.billing.AddCurrency(ctx, &billingPb.AddCurrencyRequest{
		UserId:      userID,
		Currency:    billingPb.CurrencyType_TICKETS,
		Amount:      int32(refund.Price),
		Reason:      "shop_purchase_cancel",
		ReferenceId: refID,
	})
	if err != nil {
		if isDuplicateRef(err) {
			bal, balErr := s.billing.GetBalance(ctx, &billingPb.GetBalanceRequest{
				UserId:   userID,
				Currency: billingPb.CurrencyType_TICKETS,
			})
			if balErr != nil || bal == nil {
				return 0, nil
			}
			return bal.GetBalance(), nil
		}
		return 0, err
	}
	if addResp == nil {
		return 0, nil
	}
	return addResp.GetNewBalance(), nil
}

// isDuplicateRef is true when Postgres or billing already stored this reference_id (SQLSTATE 23505).
func isDuplicateRef(err error) bool {
	if err == nil {
		return false
	}
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "duplicate") || strings.Contains(s, "23505") || strings.Contains(s, "unique constraint")
}

func (s *shopService) GetInventory(ctx context.Context, userID string) ([]repository.Item, error) {
	items, err := s.pgRepo.GetUserInventory(ctx, userID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	for i := range items {
		items[i].CanCancel = canCancelInventoryItem(items[i], now)
	}
	return items, nil
}

// canCancelInventoryItem is the player-facing refund window.
// Physical goods close once fulfillment has set fulfilled_at.
func canCancelInventoryItem(item repository.Item, now time.Time) bool {
	if item.PurchaseID == "" || item.RefundableUntil == nil {
		return false
	}
	if now.After(item.RefundableUntil.UTC()) {
		return false
	}
	if item.FulfilledAt != nil && model.IsPhysicalItem(item.Category) {
		return false
	}
	return true
}

func (s *shopService) CancelPurchase(ctx context.Context, userID, itemID string) (*CancelResult, error) {
	item, err := s.pgRepo.GetItemByID(ctx, itemID)
	if err != nil {
		// Still allow refund if purchase exists but shop item row is gone.
		item = &repository.Item{ID: itemID, Category: "merch"}
	}
	_ = item

	refund, err := s.pgRepo.RefundPurchase(ctx, userID, itemID)
	if err != nil {
		if errors.Is(err, model.ErrRefundWindowExpired) || errors.Is(err, model.ErrAlreadyFulfilled) {
			return nil, err
		}
		if strings.Contains(err.Error(), "purchase not found") {
			return nil, model.ErrPurchaseNotFound
		}
		return nil, err
	}
	if !refund.AlreadyRefunded {
		if bumpErr := s.pgRepo.BumpSpendCycle(ctx, userID, itemID); bumpErr != nil {
			log.Printf("shop: bump spend cycle user=%s item=%s: %v", userID, itemID, bumpErr)
		}
	}

	newBal, err := s.creditCancel(ctx, userID, refund)
	if err != nil {
		log.Printf("CRITICAL: purchase refunded in shop but AddCurrency failed user=%s item=%s purchase=%s: %v",
			userID, itemID, refund.PurchaseID, err)
		return nil, fmt.Errorf("failed to refund tickets: %w", err)
	}
	if refund.AlreadyRefunded {
		return &CancelResult{
			NewBalance:      newBal,
			AlreadyRefunded: true,
		}, nil
	}

	gameID := ""
	if item.GameID != nil {
		gameID = *item.GameID
	}
	_ = s.redisRepo.Delete(ctx, fmt.Sprintf("shop:items:%s:%s", item.Category, gameID))
	_ = s.redisRepo.Delete(ctx, fmt.Sprintf("shop:items:%s:", item.Category))
	_ = s.redisRepo.Delete(ctx, "shop:items:all:")
	_ = s.redisRepo.Delete(ctx, fmt.Sprintf("balance:%s:tickets", userID))

	return &CancelResult{
		NewBalance:     newBal,
		RefundedAmount: int32(refund.Price),
	}, nil
}

// checkMerchAllowed asks payment whether this player may buy a physical item.
// Without an active subscription the purchase stops before any tickets are spent.
func (s *shopService) checkMerchAllowed(ctx context.Context, userID string) error {
	if s.payment == nil {
		return fmt.Errorf("merch purchases require payment service")
	}
	gate, err := s.payment.CanPurchaseMerch(ctx, &paymentPb.CanPurchaseMerchRequest{UserId: userID})
	if err != nil {
		return fmt.Errorf("subscription check failed: %w", err)
	}
	if !gate.GetAllowed() {
		return model.ErrSubscriptionRequired
	}
	return nil
}

func (s *shopService) ListPurchasesByItemIDs(ctx context.Context, itemIDs []string, limit, offset int) ([]repository.PurchaseRecord, int64, int64, int64, error) {
	return s.pgRepo.ListPurchasesByItemIDs(ctx, itemIDs, limit, offset)
}

// SetKafkaProducer attaches an optional Kafka producer for PurchasePaid events (Week 5).
func (s *shopService) SetKafkaProducer(p kafka.Producer) {
	s.kafkaProd = p
}

func newEventID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
