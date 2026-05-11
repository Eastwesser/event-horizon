package service

import (
    "context"
    "fmt"
    "log"
    "time"

    "github.com/redis/go-redis/v9"

    "event_horizon/services/billing/internal/repository"
)

type BillingService interface {
    GetBalance(ctx context.Context, userID string, currency repository.CurrencyType) (int, error)
    AddCurrency(ctx context.Context, userID string, currency repository.CurrencyType, amount int, reason, referenceID string) (int, error)
    SpendCurrency(ctx context.Context, userID string, currency repository.CurrencyType, amount int, reason, referenceID string) (int, error)
    GetTransactionHistory(ctx context.Context, userID string, currency repository.CurrencyType, limit, offset int) ([]repository.Transaction, int, error)
}

type billingService struct {
    pgRepo  *repository.PostgresBillingRepo
    redisRepo *repository.RedisBillingRepo
}

func NewBillingService(pgRepo *repository.PostgresBillingRepo, redisRepo *repository.RedisBillingRepo) BillingService {
    return &billingService{
        pgRepo:    pgRepo,
        redisRepo: redisRepo,
    }
}

func (s *billingService) GetBalance(ctx context.Context, userID string, currency repository.CurrencyType) (int, error) {
    // Сначала пробуем получить из Redis кеша
    cached, err := s.redisRepo.GetBalance(ctx, userID, currency)
    if err != nil && err != redis.Nil {
        log.Printf("Redis get error: %v", err)
    }
    if cached >= 0 {
        return cached, nil
    }

    // Если нет в кеше — идём в PostgreSQL
    balance, err := s.pgRepo.GetBalance(ctx, userID, currency)
    if err != nil {
        return 0, err
    }

    // Сохраняем в кеш на 5 минут
    s.redisRepo.SetBalance(ctx, userID, currency, balance, 5*time.Minute)

    return balance, nil
}

func (s *billingService) AddCurrency(ctx context.Context, userID string, currency repository.CurrencyType, amount int, reason, referenceID string) (int, error) {
    if amount <= 0 {
        return 0, fmt.Errorf("amount must be positive")
    }

    // Обновляем баланс в PostgreSQL
    newBalance, err := s.pgRepo.AddBalance(ctx, userID, currency, amount, reason, referenceID)
    if err != nil {
        return 0, err
    }

    // Инвалидируем кеш
    s.redisRepo.DeleteBalance(ctx, userID, currency)

    return newBalance, nil
}

func (s *billingService) SpendCurrency(ctx context.Context, userID string, currency repository.CurrencyType, amount int, reason, referenceID string) (int, error) {
    if amount <= 0 {
        return 0, fmt.Errorf("amount must be positive")
    }

    newBalance, err := s.pgRepo.SpendBalance(ctx, userID, currency, amount, reason, referenceID)
    if err != nil {
        return 0, err
    }

    // Инвалидируем кеш
    s.redisRepo.DeleteBalance(ctx, userID, currency)

    return newBalance, nil
}

func (s *billingService) GetTransactionHistory(ctx context.Context, userID string, currency repository.CurrencyType, limit, offset int) ([]repository.Transaction, int, error) {
    if limit <= 0 || limit > 100 {
        limit = 20
    }
    return s.pgRepo.GetTransactionHistory(ctx, userID, currency, limit, offset)
}
