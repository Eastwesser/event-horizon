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