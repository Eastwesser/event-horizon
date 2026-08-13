package converter

import (
	"testing"
	"time"

	pb "github.com/Eastwesser/event-horizon/services/inventory/proto"
)

func TestItemRoundTrip(t *testing.T) {
	req := &pb.CreateItemRequest{AuthorId: "a1", Type: "брелок", Name: "Star", Price: 10.5, Stock: 3}
	item := ItemFromCreateRequest(req)
	if item.AuthorID != "a1" {
		t.Fatal(item.AuthorID)
	}
	item.ID = "id-1"
	item.CreatedAt = time.Unix(0, 0).UTC()
	item.UpdatedAt = item.CreatedAt
	proto, err := ItemToProto(item)
	if err != nil {
		t.Fatal(err)
	}
	if proto.Id != "id-1" || proto.Name != "Star" || proto.Stock != 3 {
		t.Fatalf("%+v", proto)
	}
}
