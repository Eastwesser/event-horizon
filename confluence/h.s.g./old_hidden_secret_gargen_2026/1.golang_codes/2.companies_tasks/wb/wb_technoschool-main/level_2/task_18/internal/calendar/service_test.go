package calendar

import (
	"testing"
	"time"
)

func TestServiceCRUD(t *testing.T) {
	s := NewService()
	d := time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC)
	e := Event{UserID: 1, Date: d, Text: "party"}
	if err := s.Create(e); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := s.Create(e); err == nil {
		t.Fatalf("expected duplicate error")
	}
	got := s.EventsForDay(1, d)
	if len(got) != 1 || got[0].Text != "party" {
		t.Fatalf("events: %+v", got)
	}
	if err := s.Update(1, d, "party", "afterparty"); err != nil {
		t.Fatalf("update: %v", err)
	}
	got = s.EventsForDay(1, d)
	if len(got) != 1 || got[0].Text != "afterparty" {
		t.Fatalf("updated: %+v", got)
	}
	if err := s.Delete(1, d, "afterparty"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	got = s.EventsForDay(1, d)
	if len(got) != 0 {
		t.Fatalf("expected empty, got %+v", got)
	}
}
