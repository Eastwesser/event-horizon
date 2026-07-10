package handler

import (
    "context"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    pb "github.com/Eastwesser/event-horizon/services/shop/proto"
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
        pbItems[i] = &pb.Item{
            Id:          item.ID,
            Name:        item.Name,
            Description: item.Description,
            Price:       int32(item.Price),
            Category:    item.Category,
            GameId:      item.GameID,
            ImageUrl:    item.ImageURL,
            Available:   item.Available,
            Owned:       item.Owned,
        }
    }

    return &pb.GetItemsResponse{Items: pbItems}, nil
}

func (h *ShopHandler) PurchaseItem(ctx context.Context, req *pb.PurchaseItemRequest) (*pb.PurchaseItemResponse, error) {
    newBalance, err := h.shopService.PurchaseItem(ctx, req.UserId, req.ItemId)
    if err != nil {
        return &pb.PurchaseItemResponse{
            Success: false,
            Message: err.Error(),
        }, nil
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
        pbItems[i] = &pb.Item{
            Id:          item.ID,
            Name:        item.Name,
            Description: item.Description,
            Price:       int32(item.Price),
            Category:    item.Category,
            GameId:      item.GameID,
            ImageUrl:    item.ImageURL,
            Available:   item.Available,
        }
    }

    return &pb.GetInventoryResponse{Items: pbItems}, nil
}
