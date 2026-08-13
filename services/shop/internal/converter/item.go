package converter

import (
	"github.com/Eastwesser/event-horizon/services/shop/internal/model"
	pb "github.com/Eastwesser/event-horizon/services/shop/proto"
)

func ItemToProto(item model.Item) *pb.Item {
	return &pb.Item{
		Id:          item.ID,
		Name:        item.Name,
		Description: item.Description,
		Price:       item.Price,
		Category:    item.Category,
		GameId:      item.GameID,
		ImageUrl:    item.ImageURL,
		Available:   item.Available,
	}
}

func ItemsToProto(items []model.Item) []*pb.Item {
	out := make([]*pb.Item, len(items))
	for i, it := range items {
		out[i] = ItemToProto(it)
	}
	return out
}
