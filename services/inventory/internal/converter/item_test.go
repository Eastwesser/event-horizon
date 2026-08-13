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
	item.Version = 4
	item.CreatedAt = time.Unix(0, 0).UTC()
	item.UpdatedAt = item.CreatedAt
	proto, err := ItemToProto(item)
	if err != nil {
		t.Fatal(err)
	}
	if proto.Id != "id-1" || proto.Name != "Star" || proto.Stock != 3 || proto.Version != 4 {
		t.Fatalf("%+v", proto)
	}
}

func TestItemFromUpdateRequest_Version(t *testing.T) {
	item := ItemFromUpdateRequest(&pb.UpdateItemRequest{
		Id: "x", Name: "n", Price: 1, Stock: 2, Version: 7,
	})
	if item.Version != 7 || item.ID != "x" {
		t.Fatalf("%+v", item)
	}
}
