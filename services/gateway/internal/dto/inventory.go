// Package dto maps gRPC/proto responses to stable public JSON (gin.H).
// Proto3 omitempty drops zeros and empty slices; public HTTP must keep them.
package dto

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/types/known/structpb"

	inventoryPb "github.com/Eastwesser/event-horizon/services/inventory/proto"
)

// InventoryItem maps a proto Item to gin.H with explicit zeros and [].
// Nil item → nil (caller wraps as {"item": null} or 404).
func InventoryItem(item *inventoryPb.Item) gin.H {
	if item == nil {
		return nil
	}
	images := item.GetImages()
	if images == nil {
		images = []string{}
	}
	attrs := attributesMap(item.GetAttributes())
	return gin.H{
		"id":          item.GetId(),
		"author_id":   item.GetAuthorId(),
		"type":        item.GetType(),
		"name":        item.GetName(),
		"description": item.GetDescription(),
		"price":       item.GetPrice(),
		"stock":       item.GetStock(),
		"attributes":  attrs,
		"images":      images,
		"created_at":  item.GetCreatedAt(),
		"updated_at":  item.GetUpdatedAt(),
		"version":     item.GetVersion(),
	}
}

// InventoryItemResponse wraps a single item the way FE expects: {"item": {...}}.
func InventoryItemResponse(resp *inventoryPb.ItemResponse) gin.H {
	if resp == nil {
		return gin.H{"item": nil}
	}
	return gin.H{"item": InventoryItem(resp.GetItem())}
}

// InventoryItems maps a slice; nil → empty [] for JSON.
func InventoryItems(items []*inventoryPb.Item) []gin.H {
	if items == nil {
		return []gin.H{}
	}
	out := make([]gin.H, 0, len(items))
	for _, it := range items {
		if h := InventoryItem(it); h != nil {
			out = append(out, h)
		}
	}
	return out
}

// catalogAttrOmit are bulky / internal keys not needed for grid filters.
var catalogAttrOmit = map[string]struct{}{
	"card_text":       {},
	"flavor_text":     {},
	"noiz_review":     {},
	"idempotency_key": {},
	"market_rub":      {},
}

// InventoryItemsCatalog is a smaller list DTO for SearchItems:
// keeps filterable attributes + images, drops card/flavor text.
// Detail GET still uses InventoryItem (full).
func InventoryItemsCatalog(items []*inventoryPb.Item) []gin.H {
	if items == nil {
		return []gin.H{}
	}
	out := make([]gin.H, 0, len(items))
	for _, it := range items {
		h := InventoryItem(it)
		if h == nil {
			continue
		}
		if attrs, ok := h["attributes"].(map[string]interface{}); ok && len(attrs) > 0 {
			slim := make(map[string]interface{}, len(attrs))
			for k, v := range attrs {
				if _, drop := catalogAttrOmit[k]; drop {
					continue
				}
				slim[k] = v
			}
			h["attributes"] = slim
		}
		out = append(out, h)
	}
	return out
}

func attributesMap(s *structpb.Struct) map[string]interface{} {
	if s == nil {
		return map[string]interface{}{}
	}
	return s.AsMap()
}
