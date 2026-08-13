package validate_test

import (
	"testing"

	historypb "github.com/Eastwesser/event-horizon/services/history/proto"
)

func TestRecordEventRequest_Validate(t *testing.T) {
	if err := (&historypb.RecordEventRequest{}).Validate(); err == nil {
		t.Fatal("expected error")
	}
	if err := (&historypb.RecordEventRequest{EventType: "purchase.completed"}).Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestListEventsRequest_Validate(t *testing.T) {
	if err := (&historypb.ListEventsRequest{}).Validate(); err != nil {
		t.Fatal(err)
	}
}
