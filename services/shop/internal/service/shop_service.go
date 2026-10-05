package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/Eastwesser/event-horizon/contracts/events"
	"github.com/Eastwesser/event-horizon/platform/pkg/kafka"
	"github.com/Eastwesser/event-horizon/platform/pkg/metrics"
	billingPb "github.com/Eastwesser/event-horizon/services/billing/proto"
	paymentPb "github.com/Eastwesser/event-horizon/services/payment/proto"
	"github.com/Eastwesser/event-horizon/services/shop/internal/model"
	"github.com/Eastwesser/event-horizon/services/shop/internal/repository"
	"github.com/nats-io/nats.go"
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

// Note: kafka is replaced with NATS in our project
type shopService struct {
	pgRepo    *repository.PostgresShopRepo
	redisRepo *repository.RedisShopRepo
	js        nats.JetStreamContext
	billing   billingPb.BillingServiceClient
	payment   paymentPb.PaymentServiceClient
	kafkaProd kafka.Producer // optional (noop if unset / KAFKA_BROKERS empty)
}

func NewShopService(
	pg *repository.PostgresShopRepo,
	redis *repository.RedisShopRepo,
	js nats.JetStreamContext,
	billingAddr string,
	paymentAddr string,
) ShopService {
	conn, err := grpc.Dial(billingAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("❌ Failed to connect to Billing: %v", err)
		return nil
	}
	billingClient := billingPb.NewBillingServiceClient(conn)

	var paymentClient paymentPb.PaymentServiceClient
	if paymentAddr != "" {
		pconn, perr := grpc.Dial(paymentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if perr != nil {
			log.Printf("⚠️ Failed to connect to Payment (merch gate degraded): %v", perr)
		} else {
			paymentClient = paymentPb.NewPaymentServiceClient(pconn)
		}
	}

	log.Println("🔍 Shop: trying to subscribe to inventory.item.created...")

	// ИСПРАВЛЕНО: убрал :=, оставил =
	_, err = js.Subscribe("inventory.item.created", func(msg *nats.Msg) {
		log.Println("📩 Shop: received inventory.item.created event!")

		var event map[string]interface{}
		if err := json.Unmarshal(msg.Data, &event); err != nil {
			log.Printf("Failed to parse inventory event: %v", err)
			msg.Nak()
			return
		}

		itemID, ok := event["item_id"].(string)
		if !ok || itemID == "" {
			log.Printf("❌ Invalid or empty item_id in event: %v", event)
			msg.Nak()
			return
		}

		name, _ := event["name"].(string)
		description, _ := event["description"].(string)
		price, _ := event["price"].(float64)
		stock := 1
		switch v := event["stock"].(type) {
		case float64:
			stock = int(v)
		case int:
			stock = v
		case int32:
			stock = int(v)
		case int64:
			stock = int(v)
		}

		log.Printf("📦 Creating shop item: %s (ID: %s, stock=%d)", name, itemID, stock)

		if err := pg.CreateItemFromInventory(context.Background(), itemID, name, description, price, stock); err != nil {
			log.Printf("Failed to create shop item from inventory: %v", err)
			msg.Nak()
			return
		}

		log.Printf("✅ Shop item created from inventory: %s (%s)", name, itemID)
		msg.Ack()
	}, nats.Durable("shop-inventory-sync"))

	if err != nil {
		log.Printf("⚠️ Failed to subscribe to inventory events: %v", err)
	} else {
		log.Println("✅ Shop successfully subscribed to inventory.item.created")
	}

	return &shopService{
		pgRepo:    pg,
		redisRepo: redis,
		js:        js,
		billing:   billingClient,
		payment:   paymentClient,
	}
}

func (s *shopService) GetItems(ctx context.Context, category, gameID, userID string) ([]repository.Item, error) {
	cacheKey := fmt.Sprintf("shop:items:%s:%s", category, gameID)

	// Пытаемся получить из Redis
	items, err := s.redisRepo.GetItems(ctx, cacheKey)
	if err == nil {
		// Проверяем owned для каждого товара
		for i := range items {
			owned, _ := s.pgRepo.IsItemOwned(ctx, userID, items[i].ID)
			items[i].Owned = owned
		}
		return items, nil
	}

	// Если нет в кеше — из PostgreSQL
	items, err = s.pgRepo.GetItems(ctx, category, gameID)
	if err != nil {
		return nil, err
	}

	// Сохраняем в Redis (TTL 5 минут)
	_ = s.redisRepo.SetItems(ctx, cacheKey, items, 5*time.Minute)

	// Проверяем owned
	for i := range items {
		owned, _ := s.pgRepo.IsItemOwned(ctx, userID, items[i].ID)
		items[i].Owned = owned
	}

	return items, nil
}

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

	refItem := itemID
	if len(refItem) > 8 {
		refItem = refItem[:8]
	}
	spendResp, err := s.billing.SpendCurrency(ctx, &billingPb.SpendCurrencyRequest{
		UserId:      userID,
		Currency:    billingPb.CurrencyType_TICKETS,
		Amount:      int32(item.Price),
		Reason:      "shop_purchase",
		// billing.transactions.reference_id is varchar(100); full user+item+nano overflows → silent free purchase.
		ReferenceId: fmt.Sprintf("shop-%s-%d", refItem, time.Now().UnixNano()),
		CheckOnly:   false,
	})
	if err != nil {
		if strings.Contains(err.Error(), "insufficient balance") {
			return 0, model.ErrInsufficientFunds
		}
		return 0, fmt.Errorf("failed to spend tickets: %w", err)
	}
	if spendResp == nil || !spendResp.GetSuccess() {
		msg := "spend rejected"
		if spendResp != nil && spendResp.GetMessage() != "" {
			msg = spendResp.GetMessage()
		}
		if strings.Contains(msg, "insufficient balance") {
			return 0, model.ErrInsufficientFunds
		}
		return 0, fmt.Errorf("failed to spend tickets: %s", msg)
	}

	event := map[string]interface{}{
		"user_id":   userID,
		"item_id":   itemID,
		"item_name": item.Name,
		"price":     item.Price,
		"category":  item.Category,
		"timestamp": time.Now().Unix(),
	}
	eventData, _ := json.Marshal(event)

	if err := s.pgRepo.PurchaseItemWithStock(ctx, userID, itemID, item.Price, &repository.OutboxRecord{
		EventType: "shop.purchased",
		Payload:   eventData,
	}); err != nil {
		log.Printf("CRITICAL: tickets spent but purchase failed for user %s, item %s: %v — refunding", userID, itemID, err)
		if _, refundErr := s.billing.AddCurrency(ctx, &billingPb.AddCurrencyRequest{
			UserId:      userID,
			Currency:    billingPb.CurrencyType_TICKETS,
			Amount:      int32(item.Price),
			Reason:      "shop_purchase_refund",
			ReferenceId: fmt.Sprintf("refund-%s-%d", refItem, time.Now().UnixNano()),
		}); refundErr != nil {
			log.Printf("CRITICAL: refund also failed for user %s, item %s: %v", userID, itemID, refundErr)
			event := map[string]interface{}{
				"user_id":   userID,
				"item_id":   itemID,
				"price":     item.Price,
				"error":     err.Error(),
				"refund":    refundErr.Error(),
				"timestamp": time.Now().Unix(),
			}
			eventData, _ := json.Marshal(event)
			s.js.Publish("shop.purchase.failed", eventData)
		}
		return 0, fmt.Errorf("failed to record purchase: %w", err)
	}

	if s.js != nil || s.kafkaProd != nil {
		paid := events.PurchasePaid{
			EventUUID:    newEventID(),
			PurchaseUUID: newEventID(),
			UserUUID:     userID,
			ItemUUID:     itemID,
			Price:        int32(item.Price),
		}
		if body, mErr := paid.Marshal(); mErr == nil {
			if s.js != nil {
				if _, err := s.js.Publish(kafka.TopicPurchasePaid, body); err != nil {
					log.Printf("⚠️ Failed to publish NATS purchase.paid: %v", err)
				}
			}
			if s.kafkaProd != nil {
				if err := s.kafkaProd.Send(ctx, []byte(paid.PurchaseUUID), body); err != nil {
					log.Printf("⚠️ Failed to publish kafka purchase.paid: %v", err)
				}
			}
		}
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

	return spendResp.NewBalance, nil
}

func (s *shopService) GetInventory(ctx context.Context, userID string) ([]repository.Item, error) {
	return s.pgRepo.GetUserInventory(ctx, userID)
}

func (s *shopService) CancelPurchase(ctx context.Context, userID, itemID string) (*CancelResult, error) {
	item, err := s.pgRepo.GetItemByID(ctx, itemID)
	if err != nil {
		// Still allow refund if purchase exists but shop item row is gone.
		item = &repository.Item{ID: itemID, Category: "merch"}
	}

	refund, err := s.pgRepo.RefundPurchase(ctx, userID, itemID)
	if err != nil {
		if strings.Contains(err.Error(), "purchase not found") {
			return nil, model.ErrPurchaseNotFound
		}
		return nil, err
	}
	if refund.AlreadyRefunded {
		bal, _ := s.billing.GetBalance(ctx, &billingPb.GetBalanceRequest{
			UserId:   userID,
			Currency: billingPb.CurrencyType_TICKETS,
		})
		var newBal int32
		if bal != nil {
			newBal = bal.GetBalance()
		}
		return &CancelResult{
			NewBalance:      newBal,
			AlreadyRefunded: true,
		}, nil
	}

	refID := fmt.Sprintf("cancel-%s", refund.PurchaseID)
	if len(refID) > 100 {
		refID = refID[:100]
	}
	addResp, err := s.billing.AddCurrency(ctx, &billingPb.AddCurrencyRequest{
		UserId:      userID,
		Currency:    billingPb.CurrencyType_TICKETS,
		Amount:      int32(refund.Price),
		Reason:      "shop_purchase_cancel",
		ReferenceId: refID,
	})
	if err != nil {
		log.Printf("CRITICAL: purchase refunded in shop but AddCurrency failed user=%s item=%s purchase=%s: %v",
			userID, itemID, refund.PurchaseID, err)
		return nil, fmt.Errorf("failed to refund tickets: %w", err)
	}

	gameID := ""
	if item.GameID != nil {
		gameID = *item.GameID
	}
	_ = s.redisRepo.Delete(ctx, fmt.Sprintf("shop:items:%s:%s", item.Category, gameID))
	_ = s.redisRepo.Delete(ctx, fmt.Sprintf("shop:items:%s:", item.Category))
	_ = s.redisRepo.Delete(ctx, "shop:items:all:")
	_ = s.redisRepo.Delete(ctx, fmt.Sprintf("balance:%s:tickets", userID))

	newBal := int32(0)
	if addResp != nil {
		newBal = addResp.GetNewBalance()
	}
	return &CancelResult{
		NewBalance:     newBal,
		RefundedAmount: int32(refund.Price),
	}, nil
}

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
