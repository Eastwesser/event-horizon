package converter

import (
	"fmt"

	"google.golang.org/protobuf/types/known/structpb"

	"github.com/Eastwesser/event-horizon/services/inventory/internal/model"
	pb "github.com/Eastwesser/event-horizon/services/inventory/proto"
)

func ItemFromCreateRequest(req *pb.CreateItemRequest) *model.Item {
	if req == nil {
		return nil
	}
	attrs := map[string]interface{}{}
	if req.Attributes != nil {
		attrs = req.Attributes.AsMap()
	}
	return &model.Item{
		AuthorID:    req.AuthorId,
		Type:        req.Type,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       int(req.Stock),
		Attributes:  attrs,
		Images:      req.Images,
	}
}

func ItemFromUpdateRequest(req *pb.UpdateItemRequest) *model.Item {
	if req == nil {
		return nil
	}
	attrs := map[string]interface{}{}
	if req.Attributes != nil {
		attrs = req.Attributes.AsMap()
	}
	return &model.Item{
		ID:          req.Id,
		AuthorID:    req.AuthorId,
		Type:        req.Type,
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       int(req.Stock),
		Attributes:  attrs,
		Images:      req.Images,
		Version:     int(req.Version),
	}
}

func ItemToProto(item *model.Item) (*pb.Item, error) {
	if item == nil {
		return nil, fmt.Errorf("item is nil")
	}
	var attrs *structpb.Struct
	var err error
	if item.Attributes != nil {
		attrs, err = structpb.NewStruct(item.Attributes)
		if err != nil {
			return nil, err
		}
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
		CreatedAt:   item.CreatedAt.Format(timeRFC3339),
		UpdatedAt:   item.UpdatedAt.Format(timeRFC3339),
		Version:     int32(item.Version),
	}, nil
}

const timeRFC3339 = "2006-01-02T15:04:05Z07:00"
