package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Eastwesser/event-horizon/services/leaderboard/internal/repository"
)

type mockLB struct {
	top          []repository.ScoreEntry
	topLimit     int
	rank         int
	score        int
	updateRank   int
	lastGame     string
	lastUser     string
	updateCalled bool
	err          error
}

func (m *mockLB) GetTopScores(_ context.Context, gameID string, limit, _ int) ([]repository.ScoreEntry, error) {
	m.lastGame = gameID
	m.topLimit = limit
	return m.top, m.err
}
func (m *mockLB) GetPlayerRank(_ context.Context, gameID, userID string, _ int) (int, int, error) {
	m.lastGame, m.lastUser = gameID, userID
	return m.rank, m.score, m.err
}
func (m *mockLB) UpdateScore(_ context.Context, gameID, userID, _, _ string, score, _ int) (int, error) {
	m.updateCalled = true
	m.lastGame, m.lastUser = gameID, userID
	m.score = score
	return m.updateRank, m.err
}
func (m *mockLB) UpdateScoreOnly(context.Context, string, string, string, int, int) error {
	return m.err
}
func (m *mockLB) SaveUserInfo(context.Context, string, string, string, string, int) error {
	return m.err
}

func TestGetTopScores_ClampsLimit(t *testing.T) {
	repo := &mockLB{top: []repository.ScoreEntry{{UserID: "u1", Score: 10}}}
	svc := NewLeaderboardService(repo)
	got, err := svc.GetTopScores(context.Background(), "flappy", 0, 1)
	if err != nil || len(got) != 1 || repo.topLimit != 10 {
		t.Fatalf("got=%v limit=%d err=%v", got, repo.topLimit, err)
	}
	_, _ = svc.GetTopScores(context.Background(), "flappy", 500, 1)
	if repo.topLimit != 10 {
		t.Fatalf("want clamp to 10, got %d", repo.topLimit)
	}
	_, _ = svc.GetTopScores(context.Background(), "flappy", 25, 1)
	if repo.topLimit != 25 {
		t.Fatalf("want 25, got %d", repo.topLimit)
	}
}

func TestGetPlayerRank_Delegates(t *testing.T) {
	repo := &mockLB{rank: 3, score: 900}
	rank, score, err := NewLeaderboardService(repo).GetPlayerRank(context.Background(), "hexagon", "u1", 1)
	if err != nil || rank != 3 || score != 900 || repo.lastUser != "u1" {
		t.Fatalf("rank=%d score=%d err=%v", rank, score, err)
	}
}

func TestUpdateScore_Validation(t *testing.T) {
	svc := NewLeaderboardService(&mockLB{})
	if _, err := svc.UpdateScore(context.Background(), "", "u1", "", "", 1, 1); err == nil {
		t.Fatal("expected error")
	}
	if _, err := svc.UpdateScore(context.Background(), "g", "", "", "", 1, 1); err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateScore_Success(t *testing.T) {
	repo := &mockLB{updateRank: 2}
	rank, err := NewLeaderboardService(repo).UpdateScore(context.Background(), "flappy", "u1", "a@b.c", "n", 50, 1)
	if err != nil || rank != 2 || !repo.updateCalled {
		t.Fatalf("rank=%d err=%v called=%v", rank, err, repo.updateCalled)
	}
}

func TestUpdateScoreOnlyAndSaveUserInfo(t *testing.T) {
	repo := &mockLB{}
	svc := NewLeaderboardService(repo)
	if err := svc.UpdateScoreOnly(context.Background(), "g", "u", "e", 1, 1); err != nil {
		t.Fatal(err)
	}
	if err := svc.SaveUserInfo(context.Background(), "g", "u", "e", "n", 1); err != nil {
		t.Fatal(err)
	}
	repo.err = errors.New("fail")
	if err := svc.SaveUserInfo(context.Background(), "g", "u", "e", "n", 1); err == nil {
		t.Fatal("expected error")
	}
}
