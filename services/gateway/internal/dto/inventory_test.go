package dto

import (
	"encoding/json"
	"testing"

	"google.golang.org/protobuf/types/known/structpb"

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

func TestInventoryItemsCatalogOmitsHeavyAttrs(t *testing.T) {
	attrs, err := structpb.NewStruct(map[string]any{
		"element":         "woods",
		"card_text":       "long rules text",
		"flavor_text":     "flavor",
		"idempotency_key": "k",
		"market_rub":      100,
		"hp":              float64(5),
	})
	if err != nil {
		t.Fatal(err)
	}
	item := &inventoryPb.Item{
		Id:         "a",
		Name:       "x",
		Attributes: attrs,
	}
	out := InventoryItemsCatalog([]*inventoryPb.Item{item})
	if len(out) != 1 {
		t.Fatalf("len=%d", len(out))
	}
	am, ok := out[0]["attributes"].(map[string]any)
	if !ok {
		t.Fatalf("attrs type %T", out[0]["attributes"])
	}
	if _, ok := am["card_text"]; ok {
		t.Fatal("card_text should be omitted")
	}
	if _, ok := am["flavor_text"]; ok {
		t.Fatal("flavor_text should be omitted")
	}
	if _, ok := am["idempotency_key"]; ok {
		t.Fatal("idempotency_key should be omitted")
	}
	if _, ok := am["market_rub"]; ok {
		t.Fatal("market_rub should be omitted")
	}
	if am["element"] != "woods" {
		t.Fatalf("element=%v", am["element"])
	}
	if am["hp"] != float64(5) && am["hp"] != 5 {
		// structpb may decode numbers as float64
		if v, ok := am["hp"].(float64); !ok || v != 5 {
			t.Fatalf("hp=%v (%T)", am["hp"], am["hp"])
		}
	}
}
