package handler

import (
    "context"
    "errors"
    "time"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    pb "github.com/Eastwesser/event-horizon/services/shop/proto"
    "github.com/Eastwesser/event-horizon/services/shop/internal/model"
    "github.com/Eastwesser/event-horizon/services/shop/internal/service"
)

type ShopHandler struct {
    pb.UnimplementedShopServiceServer
    shopService service.ShopService
}

func NewShopHandler(svc service.ShopService) *ShopHandler {
    return &ShopHandler{shopService: svc}
}

func (h *ShopHandler) GetItems(ctx context.Context, req *pb.GetItemsRequest) (*pb.GetItemsResponse, error) {
    items, err := h.shopService.GetItems(ctx, req.Category, req.GameId, req.UserId)
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }

    pbItems := make([]*pb.Item, len(items))
    for i, item := range items {
        gameID := ""
        if item.GameID != nil {
            gameID = *item.GameID
        }
        
        // Конвертируем PurchasedAt в строку
        purchasedAt := ""
        if item.PurchasedAt != nil {
            purchasedAt = item.PurchasedAt.Format(time.RFC3339)
        }
        
        pbItems[i] = &pb.Item{
            Id:          item.ID,
            Name:        item.Name,
            Description: item.Description,
            Price:       int32(item.Price),
            Category:    item.Category,
            GameId:      gameID,
            ImageUrl:    item.ImageURL,
            Available:   item.Available,
            Owned:       item.Owned,
            PurchasedAt: purchasedAt,
        }
    }

    return &pb.GetItemsResponse{Items: pbItems}, nil
}

func (h *ShopHandler) PurchaseItem(ctx context.Context, req *pb.PurchaseItemRequest) (*pb.PurchaseItemResponse, error) {
    newBalance, err := h.shopService.PurchaseItem(ctx, req.UserId, req.ItemId)
    if err != nil {
        return nil, mapShopErr(err)
    }

    return &pb.PurchaseItemResponse{
        Success:     true,
        Message:     "Purchase successful",
        NewBalance:  newBalance,
    }, nil
}

func (h *ShopHandler) CancelPurchase(ctx context.Context, req *pb.CancelPurchaseRequest) (*pb.CancelPurchaseResponse, error) {
    result, err := h.shopService.CancelPurchase(ctx, req.UserId, req.ItemId)
    if err != nil {
        return nil, mapShopErr(err)
    }

    msg := "Purchase cancelled"
    if result.AlreadyRefunded {
        msg = "Already refunded"
    }
    return &pb.CancelPurchaseResponse{
        Success:         true,
        Message:         msg,
        NewBalance:      result.NewBalance,
        RefundedAmount:  result.RefundedAmount,
        AlreadyRefunded: result.AlreadyRefunded,
    }, nil
}

func (h *ShopHandler) GetInventory(ctx context.Context, req *pb.GetInventoryRequest) (*pb.GetInventoryResponse, error) {
    items, err := h.shopService.GetInventory(ctx, req.UserId)
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }

    pbItems := make([]*pb.Item, len(items))
    for i, item := range items {
        gameID := ""
        if item.GameID != nil {
            gameID = *item.GameID
        }
        
        // Конвертируем PurchasedAt в строку
        purchasedAt := ""
        if item.PurchasedAt != nil {
            purchasedAt = item.PurchasedAt.Format(time.RFC3339)
        }
        
        pbItems[i] = &pb.Item{
            Id:            item.ID,
            Name:          item.Name,
            Description:   item.Description,
            Price:         int32(item.Price),
            Category:      item.Category,
            GameId:        gameID,
            ImageUrl:      item.ImageURL,
            Available:     item.Available,
            Owned:         item.Owned,
            PurchasedAt:   purchasedAt,
            PurchasePrice: int32(item.PurchasePrice),
            PurchaseId:    item.PurchaseID,
        }
    }

    return &pb.GetInventoryResponse{Items: pbItems}, nil
}

func (h *ShopHandler) ListPurchasesByItemIDs(ctx context.Context, req *pb.ListPurchasesByItemIDsRequest) (*pb.ListPurchasesByItemIDsResponse, error) {
    rows, total, salesCount, tickets, err := h.shopService.ListPurchasesByItemIDs(ctx, req.GetItemIds(), int(req.GetLimit()), int(req.GetOffset()))
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }
    out := make([]*pb.PurchaseRow, 0, len(rows))
    for _, p := range rows {
        row := &pb.PurchaseRow{
            Id:          p.ID,
            UserId:      p.UserID,
            ItemId:      p.ItemID,
            Price:       int32(p.Price),
            Status:      p.Status,
            PurchasedAt: p.PurchasedAt.Format(time.RFC3339),
        }
        if p.RefundedAt != nil {
            row.RefundedAt = p.RefundedAt.Format(time.RFC3339)
        }
        out = append(out, row)
    }
    return &pb.ListPurchasesByItemIDsResponse{
        Purchases:     out,
        Total:         total,
        SalesCount:    salesCount,
        TicketsEarned: tickets,
    }, nil
}

func mapShopErr(err error) error {
    switch {
    case errors.Is(err, model.ErrSubscriptionRequired):
        return status.Error(codes.PermissionDenied, "subscription_required")
    case errors.Is(err, model.ErrItemNotFound), errors.Is(err, model.ErrPurchaseNotFound):
        return status.Error(codes.NotFound, err.Error())
    case errors.Is(err, model.ErrAlreadyOwned):
        return status.Error(codes.AlreadyExists, err.Error())
    case errors.Is(err, model.ErrInsufficientFunds), errors.Is(err, model.ErrItemUnavailable):
        return status.Error(codes.FailedPrecondition, err.Error())
    default:
        return status.Error(codes.Internal, err.Error())
    }
}
