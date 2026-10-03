package dto

import (
	"github.com/gin-gonic/gin"

	shopPb "github.com/Eastwesser/event-horizon/services/shop/proto"
)

// ShopItem maps a shop proto Item to gin.H with explicit zeros / empty strings.
// Nil item → nil.
func ShopItem(item *shopPb.Item) gin.H {
	if item == nil {
		return nil
	}
	return gin.H{
		"id":             item.GetId(),
		"name":           item.GetName(),
		"description":    item.GetDescription(),
		"price":          item.GetPrice(),
		"category":       item.GetCategory(),
		"game_id":        item.GetGameId(),
		"image_url":      item.GetImageUrl(),
		"available":      item.GetAvailable(),
		"owned":          item.GetOwned(),
		"purchased_at":   item.GetPurchasedAt(),
		"purchase_price": item.GetPurchasePrice(),
		"purchase_id":    item.GetPurchaseId(),
	}
}

// ShopItems maps a slice; nil → empty [] for JSON.
func ShopItems(items []*shopPb.Item) []gin.H {
	if items == nil {
		return []gin.H{}
	}
	out := make([]gin.H, 0, len(items))
	for _, it := range items {
		if h := ShopItem(it); h != nil {
			out = append(out, h)
		}
	}
	return out
}

// ShopItemsCatalog is a smaller list DTO for GET /api/shop/items:
// keeps grid fields, drops description and purchase-only fields.
// Shop inventory / cancel paths should use ShopItems (full) when wired.
func ShopItemsCatalog(items []*shopPb.Item) []gin.H {
	if items == nil {
		return []gin.H{}
	}
	out := make([]gin.H, 0, len(items))
	for _, it := range items {
		if it == nil {
			continue
		}
		out = append(out, gin.H{
			"id":        it.GetId(),
			"name":      it.GetName(),
			"price":     it.GetPrice(),
			"category":  it.GetCategory(),
			"game_id":   it.GetGameId(),
			"image_url": it.GetImageUrl(),
			"available": it.GetAvailable(),
			"owned":     it.GetOwned(),
		})
	}
	return out
}
