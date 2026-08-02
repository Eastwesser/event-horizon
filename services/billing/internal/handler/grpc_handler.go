package handler

import (
    "context"
    "log"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    pb "github.com/Eastwesser/event-horizon/services/billing/proto"
    "github.com/Eastwesser/event-horizon/services/billing/internal/repository"
    "github.com/Eastwesser/event-horizon/services/billing/internal/service"
)

type BillingHandler struct {
    pb.UnimplementedBillingServiceServer
    billingService service.BillingService
}

func NewBillingHandler(svc service.BillingService) *BillingHandler {
    return &BillingHandler{
        billingService: svc,
    }
}

func (h *BillingHandler) GetBalance(ctx context.Context, req *pb.GetBalanceRequest) (*pb.GetBalanceResponse, error) {
    if req.UserId == "" {
        return nil, status.Error(codes.InvalidArgument, "user_id is required")
    }

    currency := convertProtoCurrency(req.Currency)
    balance, err := h.billingService.GetBalance(ctx, req.UserId, currency)
    if err != nil {
        log.Printf("GetBalance error: %v", err)
        return nil, status.Error(codes.Internal, err.Error())
    }

    return &pb.GetBalanceResponse{
        UserId:   req.UserId,
        Currency: req.Currency,
        Balance:  int32(balance),
    }, nil
}

func (h *BillingHandler) GetAllBalances(ctx context.Context, req *pb.GetAllBalancesRequest) (*pb.GetAllBalancesResponse, error) {
    if req.UserId == "" {
        return nil, status.Error(codes.InvalidArgument, "user_id is required")
    }

    lamps, err := h.billingService.GetBalance(ctx, req.UserId, repository.Lamps)
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }

    tickets, err := h.billingService.GetBalance(ctx, req.UserId, repository.Tickets)
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }

    return &pb.GetAllBalancesResponse{
        UserId: req.UserId,
        Balances: []*pb.BalanceEntry{
            {Currency: pb.CurrencyType_LAMPS, Balance: int32(lamps)},
            {Currency: pb.CurrencyType_TICKETS, Balance: int32(tickets)},
        },
    }, nil
}

func (h *BillingHandler) AddCurrency(ctx context.Context, req *pb.AddCurrencyRequest) (*pb.AddCurrencyResponse, error) {
    if req.UserId == "" || req.Amount <= 0 {
        return nil, status.Error(codes.InvalidArgument, "user_id and valid amount are required")
    }

    currency := convertProtoCurrency(req.Currency)
    newBalance, err := h.billingService.AddCurrency(ctx, req.UserId, currency, int(req.Amount), req.Reason, req.ReferenceId)
    if err != nil {
        log.Printf("AddCurrency error: %v", err)
        return &pb.AddCurrencyResponse{
            Success: false,
            Message: err.Error(),
        }, nil
    }

    return &pb.AddCurrencyResponse{
        Success:     true,
        NewBalance:  int32(newBalance),
        Message:     "currency added successfully",
    }, nil
}

func (h *BillingHandler) SpendCurrency(ctx context.Context, req *pb.SpendCurrencyRequest) (*pb.SpendCurrencyResponse, error) {
    if req.UserId == "" || req.Amount <= 0 {
        return nil, status.Error(codes.InvalidArgument, "user_id and valid amount are required")
    }

    currency := convertProtoCurrency(req.Currency)
    newBalance, err := h.billingService.SpendCurrency(
        ctx, 
        req.UserId, 
        currency, 
        int(req.Amount), 
        req.Reason, 
        req.ReferenceId, 
        req.CheckOnly,
    )
    if err != nil {
        log.Printf("SpendCurrency error: %v", err)
        return &pb.SpendCurrencyResponse{
            Success: false,
            Message: err.Error(),
        }, nil
    }

    return &pb.SpendCurrencyResponse{
        Success:    true,
        NewBalance: int32(newBalance),
        Message:    "currency spent successfully",
    }, nil
}

func (h *BillingHandler) GetTransactionHistory(ctx context.Context, req *pb.GetTransactionHistoryRequest) (*pb.GetTransactionHistoryResponse, error) {
    if req.UserId == "" {
        return nil, status.Error(codes.InvalidArgument, "user_id is required")
    }

    currency := convertProtoCurrency(req.Currency)
    limit := int(req.Limit)
    if limit <= 0 || limit > 100 {
        limit = 20
    }

    transactions, total, err := h.billingService.GetTransactionHistory(ctx, req.UserId, currency, limit, int(req.Offset))
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }

    pbTransactions := make([]*pb.Transaction, len(transactions))
    for i, t := range transactions {
        pbTransactions[i] = &pb.Transaction{
            Id:           t.ID,
            UserId:       t.UserID,
            Currency:     convertToProtoCurrency(t.Currency),
            Amount:       int32(t.Amount),
            BalanceAfter: int32(t.BalanceAfter),
            Reason:       t.Reason,
            ReferenceId:  t.ReferenceID,
            CreatedAt:    t.CreatedAt.Unix(),
        }
    }

    return &pb.GetTransactionHistoryResponse{
        Transactions: pbTransactions,
        Total:        int32(total),
    }, nil
}

func convertProtoCurrency(c pb.CurrencyType) repository.CurrencyType {
    switch c {
    case pb.CurrencyType_LAMPS:
        return repository.Lamps
    case pb.CurrencyType_TICKETS:
        return repository.Tickets
    default:
        return repository.Lamps
    }
}

func convertToProtoCurrency(c repository.CurrencyType) pb.CurrencyType {
    switch c {
    case repository.Lamps:
        return pb.CurrencyType_LAMPS
    case repository.Tickets:
        return pb.CurrencyType_TICKETS
    default:
        return pb.CurrencyType_LAMPS
    }
}
