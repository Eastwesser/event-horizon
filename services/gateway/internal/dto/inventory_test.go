package dto

import (
	"encoding/json"
	"testing"

	inventoryPb "github.com/Eastwesser/event-horizon/services/inventory/proto"
)

func TestInventoryItemKeepsZeros(t *testing.T) {
	item := &inventoryPb.Item{
		Id:     "a",
		Name:   "free",
		Price:  0,
		Stock:  0,
		Images: nil,
	}
	h := InventoryItem(item)
	b, err := json.Marshal(h)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}
	if _, ok := m["price"]; !ok {
		t.Fatalf("price omitted: %s", b)
	}
	if _, ok := m["stock"]; !ok {
		t.Fatalf("stock omitted: %s", b)
	}
	imgs, ok := m["images"].([]any)
	if !ok {
		t.Fatalf("images not array: %s", b)
	}
	if len(imgs) != 0 {
		t.Fatalf("want empty images, got %v", imgs)
	}
}

func TestInventoryItemsNilIsEmpty(t *testing.T) {
	out := InventoryItems(nil)
	b, _ := json.Marshal(out)
	if string(b) != "[]" {
		t.Fatalf("want [], got %s", b)
	}
}
