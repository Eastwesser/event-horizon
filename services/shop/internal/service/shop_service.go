package service

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    "github.com/Eastwesser/event-horizon/services/shop/internal/repository"
    billingPb "github.com/Eastwesser/event-horizon/services/billing/proto"
    "github.com/nats-io/nats.go"
)

type ShopService interface {
    GetItems(ctx context.Context, category, gameID, userID string) ([]repository.Item, error)
    PurchaseItem(ctx context.Context, userID, itemID string) (int32, error)
    GetInventory(ctx context.Context, userID string) ([]repository.Item, error)
}

type shopService struct {
    pgRepo    *repository.PostgresShopRepo
    redisRepo *repository.RedisShopRepo
    js        nats.JetStreamContext
    billing   billingPb.BillingServiceClient
}

func NewShopService(pg *repository.PostgresShopRepo, redis *repository.RedisShopRepo, js nats.JetStreamContext, billingAddr string) ShopService {
    conn, err := grpc.Dial(billingAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Printf("❌ Failed to connect to Billing: %v", err)
        return nil
    }
    billingClient := billingPb.NewBillingServiceClient(conn)

    return &shopService{
        pgRepo:    pg,
        redisRepo: redis,
        js:        js,
        billing:   billingClient,
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
    // 1. Проверяем, существует ли товар
    item, err := s.pgRepo.GetItemByID(ctx, itemID)
    if err != nil {
        return 0, fmt.Errorf("item not found: %w", err)
    }
    if !item.Available {
        return 0, fmt.Errorf("item is not available")
    }

    // 2. Проверяем, не куплен ли уже
    owned, err := s.pgRepo.IsItemOwned(ctx, userID, itemID)
    if err != nil {
        return 0, err
    }
    if owned {
        return 0, fmt.Errorf("item already purchased")
    }

    // 3. Проверяем баланс через Billing
    balanceReq := &billingPb.GetAllBalancesRequest{UserId: userID}
    balanceResp, err := s.billing.GetAllBalances(ctx, balanceReq)
    if err != nil {
        return 0, fmt.Errorf("failed to get balance: %w", err)
    }

    var ticketsBalance int32
    for _, b := range balanceResp.Balances {
        if b.Currency == billingPb.CurrencyType_TICKETS {
            ticketsBalance = b.Balance
            break
        }
    }

    if ticketsBalance < int32(item.Price) {
        return 0, fmt.Errorf("not enough tickets: have %d, need %d", ticketsBalance, item.Price)
    }

    // 4. Списываем билетики через Billing
    _, err = s.billing.SpendCurrency(ctx, &billingPb.SpendCurrencyRequest{
        UserId:   userID,
        Currency: billingPb.CurrencyType_TICKETS,
        Amount:   int32(item.Price),
        Reason:   "shop_purchase",
    })
    if err != nil {
        return 0, fmt.Errorf("failed to deduct tickets: %w", err)
    }

    // 5. Записываем покупку в БД
    if err := s.pgRepo.PurchaseItem(ctx, userID, itemID, item.Price); err != nil {
        return 0, fmt.Errorf("failed to record purchase: %w", err)
    }

    // 6. Публикуем событие в NATS
    event := map[string]interface{}{
        "user_id":   userID,
        "item_id":   itemID,
        "item_name": item.Name,
        "price":     item.Price,
        "category":  item.Category,
        "timestamp": time.Now().Unix(),
    }
    eventData, _ := json.Marshal(event)
    _, err = s.js.Publish("shop.purchased", eventData)
    if err != nil {
        log.Printf("⚠️ Failed to publish shop.purchased: %v", err)
    }

    // 7. Очищаем кеш товаров (все кеши, а не только по категории)
    _ = s.redisRepo.Delete(ctx, fmt.Sprintf("shop:items:%s:%s", item.Category, item.GameID))
    _ = s.redisRepo.Delete(ctx, fmt.Sprintf("shop:items:%s:", item.Category))
    _ = s.redisRepo.Delete(ctx, "shop:items:all:")

    // 8. Очищаем кеш баланса
    cacheKey := fmt.Sprintf("balance:%s:tickets", userID)
    if err := s.redisRepo.Delete(ctx, cacheKey); err != nil {
        log.Printf("⚠️ Failed to delete balance cache: %v", err)
    }

    // 9. Возвращаем новый баланс
    newBalance := ticketsBalance - int32(item.Price)
    return newBalance, nil
}

func (s *shopService) GetInventory(ctx context.Context, userID string) ([]repository.Item, error) {
    return s.pgRepo.GetUserInventory(ctx, userID)
}
