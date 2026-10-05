package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Eastwesser/event-horizon/services/profile/internal/repository"
)

type mockProfileRepo struct {
	profile *repository.UserProfile
	err     error
	upserts int
}

func (m *mockProfileRepo) GetProfile(context.Context, string) (*repository.UserProfile, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.profile == nil {
		return nil, nil
	}
	cp := *m.profile
	return &cp, nil
}
func (m *mockProfileRepo) UpsertProfile(_ context.Context, p *repository.UserProfile) error {
	m.upserts++
	m.profile = p
	return m.err
}

type mockAchievements struct {
	list    []repository.Achievement
	unlocks []string
	err     error
}

func (m *mockAchievements) ListUnlocked(context.Context, string) ([]repository.Achievement, error) {
	return m.list, m.err
}
func (m *mockAchievements) Unlock(_ context.Context, _ string, codes []string) (int64, error) {
	m.unlocks = append(m.unlocks, codes...)
	return int64(len(codes)), m.err
}

func TestGetProfile_FromRepo(t *testing.T) {
	repo := &mockProfileRepo{profile: &repository.UserProfile{UserID: "u1", Nickname: "neo"}}
	got, err := NewProfileService(repo, nil, nil, time.Minute).GetProfile(context.Background(), "u1")
	if err != nil || got.Nickname != "neo" {
		t.Fatalf("%+v err=%v", got, err)
	}
}

func TestUpdateProfile(t *testing.T) {
	repo := &mockProfileRepo{}
	svc := NewProfileService(repo, nil, nil, time.Minute)
	if err := svc.UpdateProfile(context.Background(), &repository.UserProfile{UserID: "u1", Nickname: "x"}); err != nil {
		t.Fatal(err)
	}
	if repo.upserts != 1 {
		t.Fatal("upsert not called")
	}
	repo.err = errors.New("db")
	if err := svc.UpdateProfile(context.Background(), &repository.UserProfile{UserID: "u1"}); err == nil {
		t.Fatal("expected error")
	}
}

func TestListAchievements(t *testing.T) {
	svc := NewProfileService(&mockProfileRepo{}, nil, nil, time.Minute)
	got, err := svc.ListAchievements(context.Background(), "u1")
	if err != nil || got != nil {
		t.Fatalf("%v err=%v", got, err)
	}
	ach := &mockAchievements{list: []repository.Achievement{{Code: "first_play_flappy"}}}
	got, err = NewProfileService(&mockProfileRepo{}, ach, nil, time.Minute).ListAchievements(context.Background(), "u1")
	if err != nil || len(got) != 1 {
		t.Fatalf("%v err=%v", got, err)
	}
}

func TestEvaluateAndUnlock(t *testing.T) {
	ach := &mockAchievements{}
	svc := NewProfileService(&mockProfileRepo{}, ach, nil, time.Minute)
	if err := svc.EvaluateAndUnlock(context.Background(), "", map[string]int32{"flappy": 1}, 1, "flappy", 1); err != nil {
		t.Fatal(err)
	}
	if len(ach.unlocks) != 0 {
		t.Fatal("empty user should no-op")
	}
	if err := svc.EvaluateAndUnlock(context.Background(), "u1", map[string]int32{"flappy": 100}, 1000, "flappy", 5); err != nil {
		t.Fatal(err)
	}
	if len(ach.unlocks) == 0 {
		t.Fatal("expected unlocks")
	}
}
