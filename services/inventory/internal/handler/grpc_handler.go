package handler

import (
    "context"
    "strconv"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/protobuf/types/known/structpb"

    "github.com/Eastwesser/event-horizon/services/inventory/internal/model"
    "github.com/Eastwesser/event-horizon/services/inventory/internal/service"
    pb "github.com/Eastwesser/event-horizon/services/inventory/proto"
)

type GRPCHandler struct {
    pb.UnimplementedInventoryServiceServer
    service *service.InventoryService
}

func NewGRPCHandler(svc *service.InventoryService) *GRPCHandler {
    return &GRPCHandler{
        service: svc,
    }
}

func itemToProto(item *model.Item) (*pb.Item, error) {
    attrs, err := structpb.NewStruct(item.Attributes)
    if err != nil {
        return nil, err
    }

    return &pb.Item{
        Id:          item.ID,
        AuthorId:    item.AuthorID,
        Type:        item.Type,
        Name:        item.Name,
        Description: item.Description,
        Price:       item.Price,
        Stock:       int32(item.Stock),
        Attributes:  attrs,
        Images:      item.Images,
        CreatedAt:   item.CreatedAt.String(),
        UpdatedAt:   item.UpdatedAt.String(),
    }, nil
}

func (h *GRPCHandler) CreateItem(ctx context.Context, req *pb.CreateItemRequest) (*pb.ItemResponse, error) {
    attrs := req.Attributes.AsMap()
    item := &model.Item{
        AuthorID:    req.AuthorId,
        Type:        req.Type,
        Name:        req.Name,
        Description: req.Description,
        Price:       req.Price,
        Stock:       int(req.Stock),
        Attributes:  attrs,
        Images:      req.Images,
    }

    if err := h.service.CreateItem(ctx, item); err != nil {
        return nil, status.Errorf(codes.Internal, "failed to create item: %v", err)
    }

    protoItem, err := itemToProto(item)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to convert item: %v", err)
    }

    return &pb.ItemResponse{Item: protoItem}, nil
}

func (h *GRPCHandler) GetItem(ctx context.Context, req *pb.GetItemRequest) (*pb.ItemResponse, error) {
    item, err := h.service.GetItem(ctx, req.Id)
    if err != nil {
        return nil, status.Errorf(codes.NotFound, "item not found: %v", err)
    }

    protoItem, err := itemToProto(item)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to convert item: %v", err)
    }

    return &pb.ItemResponse{Item: protoItem}, nil
}

func (h *GRPCHandler) UpdateItem(ctx context.Context, req *pb.UpdateItemRequest) (*pb.ItemResponse, error) {
    attrs := req.Attributes.AsMap()
    item := &model.Item{
        ID:          req.Id,
        AuthorID:    req.AuthorId,
        Type:        req.Type,
        Name:        req.Name,
        Description: req.Description,
        Price:       req.Price,
        Stock:       int(req.Stock),
        Attributes:  attrs,
        Images:      req.Images,
    }

    if err := h.service.UpdateItem(ctx, item); err != nil {
        return nil, status.Errorf(codes.Internal, "failed to update item: %v", err)
    }

    protoItem, err := itemToProto(item)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to convert item: %v", err)
    }

    return &pb.ItemResponse{Item: protoItem}, nil
}

func (h *GRPCHandler) DeleteItem(ctx context.Context, req *pb.DeleteItemRequest) (*pb.EmptyResponse, error) {
    if err := h.service.DeleteItem(ctx, req.Id); err != nil {
        return nil, status.Errorf(codes.Internal, "failed to delete item: %v", err)
    }
    return &pb.EmptyResponse{}, nil
}

func (h *GRPCHandler) SearchItems(ctx context.Context, req *pb.SearchItemsRequest) (*pb.SearchItemsResponse, error) {
    filters := make(map[string]interface{})
    for k, v := range req.Filters {
        if k == "price_min" || k == "price_max" {
            if val, err := strconv.ParseFloat(v, 64); err == nil {
                filters[k] = val
            }
        } else {
            filters[k] = v
        }
    }

    items, total, err := h.service.SearchItems(ctx, filters, int(req.Limit), int(req.Offset))
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to search items: %v", err)
    }

    protoItems := make([]*pb.Item, len(items))
    for i, item := range items {
        protoItem, err := itemToProto(item)
        if err != nil {
            return nil, status.Errorf(codes.Internal, "failed to convert item: %v", err)
        }
        protoItems[i] = protoItem
    }

    return &pb.SearchItemsResponse{
        Items: protoItems,
        Total: total,
    }, nil
}

func (h *GRPCHandler) GetByAuthor(ctx context.Context, req *pb.GetByAuthorRequest) (*pb.SearchItemsResponse, error) {
    items, total, err := h.service.GetByAuthor(ctx, req.AuthorId, int(req.Limit), int(req.Offset))
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to get items by author: %v", err)
    }

    protoItems := make([]*pb.Item, len(items))
    for i, item := range items {
        protoItem, err := itemToProto(item)
        if err != nil {
            return nil, status.Errorf(codes.Internal, "failed to convert item: %v", err)
        }
        protoItems[i] = protoItem
    }

    return &pb.SearchItemsResponse{
        Items: protoItems,
        Total: total,
    }, nil
}

func (h *GRPCHandler) GetByType(ctx context.Context, req *pb.GetByTypeRequest) (*pb.SearchItemsResponse, error) {
    items, total, err := h.service.GetByType(ctx, req.Type, int(req.Limit), int(req.Offset))
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to get items by type: %v", err)
    }

    protoItems := make([]*pb.Item, len(items))
    for i, item := range items {
        protoItem, err := itemToProto(item)
        if err != nil {
            return nil, status.Errorf(codes.Internal, "failed to convert item: %v", err)
        }
        protoItems[i] = protoItem
    }

    return &pb.SearchItemsResponse{
        Items: protoItems,
        Total: total,
    }, nil
}

// BulkCreateItems - массовое создание товаров
func (h *GRPCHandler) BulkCreateItems(ctx context.Context, req *pb.BulkCreateItemsRequest) (*pb.BulkCreateItemsResponse, error) {
    items := make([]*model.Item, len(req.Items))
    for i, pbItem := range req.Items {
        attrs := pbItem.Attributes.AsMap()
        items[i] = &model.Item{
            AuthorID:    pbItem.AuthorId,
            Type:        pbItem.Type,
            Name:        pbItem.Name,
            Description: pbItem.Description,
            Price:       pbItem.Price,
            Stock:       int(pbItem.Stock),
            Attributes:  attrs,
            Images:      pbItem.Images,
        }
    }

    if err := h.service.BulkCreateItems(ctx, items); err != nil {
        return nil, status.Errorf(codes.Internal, "failed to bulk create items: %v", err)
    }

    return &pb.BulkCreateItemsResponse{
        Success: true,
        Count:   int32(len(items)),
    }, nil
}

// ReserveItem - резервирование товара
func (h *GRPCHandler) ReserveItem(ctx context.Context, req *pb.ReserveItemRequest) (*pb.ReserveItemResponse, error) {
    remaining, err := h.service.ReserveItem(ctx, req.Id, int(req.Quantity))
    if err != nil {
        if err == model.ErrItemNotFound {
            return nil, status.Errorf(codes.NotFound, "item not found: %v", err)
        }
        if err == model.ErrNotEnoughStock {
            return nil, status.Errorf(codes.FailedPrecondition, "not enough stock: %v", err)
        }
        return nil, status.Errorf(codes.Internal, "failed to reserve item: %v", err)
    }

    return &pb.ReserveItemResponse{
        Success:        true,
        RemainingStock: int32(remaining),
    }, nil
}

// SoftDeleteItem - мягкое удаление
func (h *GRPCHandler) SoftDeleteItem(ctx context.Context, req *pb.SoftDeleteItemRequest) (*pb.EmptyResponse, error) {
    if err := h.service.SoftDeleteItem(ctx, req.Id); err != nil {
        if err == model.ErrItemNotFound {
            return nil, status.Errorf(codes.NotFound, "item not found: %v", err)
        }
        return nil, status.Errorf(codes.Internal, "failed to soft delete item: %v", err)
    }
    return &pb.EmptyResponse{}, nil
}

// RestoreItem - восстановление после мягкого удаления
func (h *GRPCHandler) RestoreItem(ctx context.Context, req *pb.RestoreItemRequest) (*pb.EmptyResponse, error) {
    if err := h.service.RestoreItem(ctx, req.Id); err != nil {
        if err == model.ErrItemNotFound {
            return nil, status.Errorf(codes.NotFound, "item not found: %v", err)
        }
        return nil, status.Errorf(codes.Internal, "failed to restore item: %v", err)
    }
    return &pb.EmptyResponse{}, nil
}

// GetStats - статистика
func (h *GRPCHandler) GetStats(ctx context.Context, req *pb.EmptyRequest) (*pb.StatsResponse, error) {
    stats, err := h.service.GetStats(ctx)
    if err != nil {
        return nil, status.Errorf(codes.Internal, "failed to get stats: %v", err)
    }

    return &pb.StatsResponse{
        TotalItems: stats.TotalItems,
        ByType:     stats.ByType,
        ByAuthor:   stats.ByAuthor,
    }, nil
}