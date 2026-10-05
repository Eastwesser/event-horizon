package service

import (
	"context"
	"log/slog"
	"testing"

	"github.com/Eastwesser/event-horizon/services/notification/internal/model"
)

type memStore struct {
	items []*model.Notification
}

func (m *memStore) Insert(_ context.Context, n *model.Notification) (string, error) {
	for _, x := range m.items {
		if x.UserID == n.UserID && x.Source == n.Source && n.Source != "" {
			return x.ID, nil
		}
	}
	if n.ID == "" {
		n.ID = "id-" + n.Source
	}
	cp := *n
	m.items = append(m.items, &cp)
	return n.ID, nil
}

func (m *memStore) List(_ context.Context, userID string, _, _ int, unreadOnly bool) ([]*model.Notification, int64, int64, error) {
	var out []*model.Notification
	var unread int64
	for _, n := range m.items {
		if n.UserID != userID {
			continue
		}
		if n.ReadAt == nil {
			unread++
		}
		if unreadOnly && n.ReadAt != nil {
			continue
		}
		out = append(out, n)
	}
	return out, int64(len(out)), unread, nil
}

func (m *memStore) MarkRead(_ context.Context, userID, id string) (int64, error) {
	var n int64
	for _, x := range m.items {
		if x.UserID != userID {
			continue
		}
		if id != "" && x.ID != id {
			continue
		}
		if x.ReadAt == nil {
			t := x.CreatedAt
			x.ReadAt = &t
			n++
		}
	}
	return n, nil
}

func TestHandleSubmitted_CreatesPerAdmin(t *testing.T) {
	st := &memStore{}
	s := NewInbox(st, slog.Default())
	if err := s.HandleSubmitted(context.Background(), "app1", "u1", "Иван", []string{"a1", "a2", ""}); err != nil {
		t.Fatal(err)
	}
	if len(st.items) != 2 {
		t.Fatalf("want 2 rows, got %d", len(st.items))
	}
}

func TestHandleApproved(t *testing.T) {
	st := &memStore{}
	s := NewInbox(st, slog.Default())
	if err := s.HandleApproved(context.Background(), "app1", "u1"); err != nil {
		t.Fatal(err)
	}
	if len(st.items) != 1 || st.items[0].Link != "/author/dashboard" {
		t.Fatalf("unexpected: %+v", st.items)
	}
}

func TestHandleSubmitted_NoAdmins(t *testing.T) {
	st := &memStore{}
	s := NewInbox(st, slog.Default())
	if err := s.HandleSubmitted(context.Background(), "app1", "u1", "X", nil); err != nil {
		t.Fatal(err)
	}
	if len(st.items) != 0 {
		t.Fatalf("expected empty")
	}
}
