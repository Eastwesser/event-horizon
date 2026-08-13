package converter

import (
	"testing"

	"github.com/Eastwesser/event-horizon/services/shop/internal/model"
)

func TestItemToProto(t *testing.T) {
	p := ItemToProto(model.Item{ID: "1", Name: "skin", Price: 100, Available: true})
	if p.Id != "1" || p.Price != 100 || !p.Available {
		t.Fatalf("%+v", p)
	}
}
