package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/Eastwesser/event-horizon/services/history/internal/model"
)

type mockHistory struct {
	lastPayload string
	lastType    string
	lastUser    string
	before      time.Time
	insertID    string
	list        []*model.Event
	purged      int64
	err         error
}

func (m *mockHistory) Insert(_ context.Context, userID, eventType, payload string) (string, error) {
	m.lastUser, m.lastType, m.lastPayload = userID, eventType, payload
	if m.err != nil {
		return "", m.err
	}
	if m.insertID == "" {
		return "evt-1", nil
	}
	return m.insertID, nil
}
func (m *mockHistory) List(_ context.Context, userID, eventType string, _, _ int) ([]*model.Event, int64, error) {
	m.lastUser, m.lastType = userID, eventType
	return m.list, int64(len(m.list)), m.err
}
func (m *mockHistory) DeleteOlderThan(_ context.Context, before time.Time) (int64, error) {
	m.before = before
	return m.purged, m.err
}

func TestNew_DefaultRetention(t *testing.T) {
	s := New(&mockHistory{}, 0)
	if s.retentionDays != 30 {
		t.Fatalf("%d", s.retentionDays)
	}
}

func TestRecordEvent_ValidationAndNormalize(t *testing.T) {
	repo := &mockHistory{}
	svc := New(repo, 30)
	if _, err := svc.RecordEvent(context.Background(), "u1", "", "{}"); !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("%v", err)
	}
	id, err := svc.RecordEvent(context.Background(), "u1", "login", "")
	if err != nil || id == "" || repo.lastPayload != "{}" {
		t.Fatalf("id=%s payload=%q err=%v", id, repo.lastPayload, err)
	}
	_, err = svc.RecordEvent(context.Background(), "u1", "x", "not-json")
	if err != nil {
		t.Fatal(err)
	}
	var wrapped map[string]string
	if err := json.Unmarshal([]byte(repo.lastPayload), &wrapped); err != nil || wrapped["raw"] != "not-json" {
		t.Fatalf("payload=%s", repo.lastPayload)
	}
}

func TestListEvents_Delegates(t *testing.T) {
	repo := &mockHistory{list: []*model.Event{{ID: "1", EventType: "login"}}}
	got, total, err := New(repo, 7).ListEvents(context.Background(), "u1", "login", 10, 0)
	if err != nil || total != 1 || got[0].ID != "1" {
		t.Fatalf("%v %d %v", got, total, err)
	}
}

func TestPurgeExpired_UsesRetention(t *testing.T) {
	repo := &mockHistory{purged: 3}
	n, err := New(repo, 7).PurgeExpired(context.Background())
	if err != nil || n != 3 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	want := time.Now().UTC().AddDate(0, 0, -7)
	if repo.before.Before(want.Add(-time.Minute)) || repo.before.After(want.Add(time.Minute)) {
		t.Fatalf("before=%v want~%v", repo.before, want)
	}
}
