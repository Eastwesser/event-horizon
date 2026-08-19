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

    return &pb.GetInventoryResponse{Items: pbItems}, nil
}

func mapShopErr(err error) error {
    switch {
    case errors.Is(err, model.ErrSubscriptionRequired):
        return status.Error(codes.PermissionDenied, "subscription_required")
    case errors.Is(err, model.ErrItemNotFound):
        return status.Error(codes.NotFound, err.Error())
    case errors.Is(err, model.ErrAlreadyOwned):
        return status.Error(codes.AlreadyExists, err.Error())
    case errors.Is(err, model.ErrInsufficientFunds), errors.Is(err, model.ErrItemUnavailable):
        return status.Error(codes.FailedPrecondition, err.Error())
    default:
        return status.Error(codes.Internal, err.Error())
    }
}
