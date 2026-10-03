package dto

import (
	"encoding/json"
	"testing"

	shopPb "github.com/Eastwesser/event-horizon/services/shop/proto"
)

func TestShopItemKeepsZeros(t *testing.T) {
	item := &shopPb.Item{
		Id:            "a",
		Name:          "free",
		Price:         0,
		Available:     false,
		Owned:         false,
		PurchasePrice: 0,
	}
	h := ShopItem(item)
	b, err := json.Marshal(h)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"price", "available", "owned", "description", "purchase_price", "purchased_at", "purchase_id"} {
		if _, ok := m[key]; !ok {
			t.Fatalf("%s omitted: %s", key, b)
		}
	}
}

func TestShopItemsNilIsEmpty(t *testing.T) {
	out := ShopItems(nil)
	b, _ := json.Marshal(out)
	if string(b) != "[]" {
		t.Fatalf("want [], got %s", b)
	}
}

func TestShopItemsCatalogOmitsHeavyFields(t *testing.T) {
	item := &shopPb.Item{
		Id:            "a",
		Name:          "skin",
		Description:   "long description text",
		Price:         0,
		Category:      "game_skin",
		GameId:        "flappy",
		ImageUrl:      "",
		Available:     true,
		Owned:         false,
		PurchasedAt:   "2026-01-01T00:00:00Z",
		PurchasePrice: 100,
		PurchaseId:    "p1",
	}
	out := ShopItemsCatalog([]*shopPb.Item{item})
	if len(out) != 1 {
		t.Fatalf("len=%d", len(out))
	}
	b, err := json.Marshal(out[0])
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"description", "purchased_at", "purchase_price", "purchase_id"} {
		if _, ok := m[key]; ok {
			t.Fatalf("%s should be omitted: %s", key, b)
		}
	}
	for _, key := range []string{"id", "name", "price", "category", "game_id", "image_url", "available", "owned"} {
		if _, ok := m[key]; !ok {
			t.Fatalf("%s missing: %s", key, b)
		}
	}
	if m["price"].(float64) != 0 {
		t.Fatalf("price zero dropped: %s", b)
	}
	if m["owned"].(bool) != false {
		t.Fatalf("owned false dropped: %s", b)
	}
}

func TestShopItemsCatalogNilIsEmpty(t *testing.T) {
	out := ShopItemsCatalog(nil)
	b, _ := json.Marshal(out)
	if string(b) != "[]" {
		t.Fatalf("want [], got %s", b)
	}
}
