package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Eastwesser/event-horizon/services/analytics/internal/model"
)

type mockAnalytics struct {
	lastPayload string
	lastType    string
	lastDays    int
	dau         []model.DayCount
	mau         int64
	ret         *model.Retention
	err         error
}

func (m *mockAnalytics) Record(_ context.Context, _, eventType, payload string) error {
	m.lastType, m.lastPayload = eventType, payload
	return m.err
}
func (m *mockAnalytics) DAU(_ context.Context, days int) ([]model.DayCount, error) {
	m.lastDays = days
	return m.dau, m.err
}
func (m *mockAnalytics) MAU(_ context.Context, days int) (int64, error) {
	m.lastDays = days
	return m.mau, m.err
}
func (m *mockAnalytics) Retention(context.Context, int, int) (*model.Retention, error) {
	return m.ret, m.err
}

func TestRecordEvent(t *testing.T) {
	repo := &mockAnalytics{}
	svc := New(repo)
	if err := svc.RecordEvent(context.Background(), "u", "", "x"); !errors.Is(err, model.ErrInvalidInput) {
		t.Fatalf("%v", err)
	}
	if err := svc.RecordEvent(context.Background(), "u", "play", ""); err != nil || repo.lastPayload != "{}" {
		t.Fatalf("payload=%q err=%v", repo.lastPayload, err)
	}
}

func TestGetMAU_ClampsDays(t *testing.T) {
	repo := &mockAnalytics{mau: 42}
	n, days, err := New(repo).GetMAU(context.Background(), 0)
	if err != nil || n != 42 || days != 30 || repo.lastDays != 30 {
		t.Fatalf("n=%d days=%d last=%d err=%v", n, days, repo.lastDays, err)
	}
	n, days, err = New(repo).GetMAU(context.Background(), 14)
	if err != nil || days != 14 || repo.lastDays != 14 {
		t.Fatalf("days=%d err=%v", days, err)
	}
}

func TestGetDAUAndRetention(t *testing.T) {
	repo := &mockAnalytics{
		dau: []model.DayCount{{Day: "2026-10-01", Count: 5}},
		ret: &model.Retention{CohortSize: 10},
	}
	dau, err := New(repo).GetDAU(context.Background(), 7)
	if err != nil || len(dau) != 1 || repo.lastDays != 7 {
		t.Fatalf("%v err=%v", dau, err)
	}
	ret, err := New(repo).GetRetention(context.Background(), 7, 7)
	if err != nil || ret.CohortSize != 10 {
		t.Fatalf("%+v err=%v", ret, err)
	}
}
