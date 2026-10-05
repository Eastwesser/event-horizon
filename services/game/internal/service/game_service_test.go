package service

import (
	"context"
	"errors"
	"testing"

	billingPb "github.com/Eastwesser/event-horizon/services/billing/proto"
	"github.com/Eastwesser/event-horizon/services/game/internal/repository"
	"google.golang.org/grpc"
)

type mockGameRepo struct {
	highscore   int
	saveHS      bool
	enqueue     bool
	saveEnqueue bool
	activeBoost string
	consumeOK   bool
	consumeID   string
	createBoost bool
	err         error
}

func (m *mockGameRepo) GetHighscore(context.Context, string, string, int) (int, error) {
	return m.highscore, m.err
}
func (m *mockGameRepo) SaveHighscore(context.Context, string, string, int, int) error {
	m.saveHS = true
	return m.err
}
func (m *mockGameRepo) EnqueueOutbox(context.Context, string, []byte) error {
	m.enqueue = true
	return m.err
}
func (m *mockGameRepo) SaveHighscoreAndEnqueueOutbox(context.Context, string, string, int, int, string, []byte) error {
	m.saveEnqueue = true
	return m.err
}
func (m *mockGameRepo) CreateRunBoost(context.Context, string, string, string) error {
	m.createBoost = true
	return m.err
}
func (m *mockGameRepo) GetActiveRunBoost(context.Context, string, string) (string, bool, error) {
	if m.activeBoost == "" {
		return "", false, nil
	}
	return m.activeBoost, true, nil
}
func (m *mockGameRepo) ConsumeActiveRunBoost(context.Context, string, string, string) (string, bool, error) {
	if m.consumeOK {
		return m.consumeID, true, nil
	}
	return "", false, nil
}

type stubBilling struct {
	spendOK  bool
	spendBal int32
	spendErr error
}

func (b *stubBilling) GetBalance(context.Context, *billingPb.GetBalanceRequest, ...grpc.CallOption) (*billingPb.GetBalanceResponse, error) {
	return nil, errors.New("unused")
}
func (b *stubBilling) GetAllBalances(context.Context, *billingPb.GetAllBalancesRequest, ...grpc.CallOption) (*billingPb.GetAllBalancesResponse, error) {
	return nil, errors.New("unused")
}
func (b *stubBilling) AddCurrency(context.Context, *billingPb.AddCurrencyRequest, ...grpc.CallOption) (*billingPb.AddCurrencyResponse, error) {
	return nil, errors.New("unused")
}
func (b *stubBilling) SpendCurrency(context.Context, *billingPb.SpendCurrencyRequest, ...grpc.CallOption) (*billingPb.SpendCurrencyResponse, error) {
	if b.spendErr != nil {
		return nil, b.spendErr
	}
	return &billingPb.SpendCurrencyResponse{Success: b.spendOK, NewBalance: b.spendBal}, nil
}
func (b *stubBilling) GetTransactionHistory(context.Context, *billingPb.GetTransactionHistoryRequest, ...grpc.CallOption) (*billingPb.GetTransactionHistoryResponse, error) {
	return nil, errors.New("unused")
}

func TestGetGameInfo(t *testing.T) {
	svc := NewGameService(&mockGameRepo{}, nil, nil)
	for _, id := range []string{"hexagon", "memory", "flappy", "towers", "hanoi", "twenty48", "gears", "companion"} {
		info, err := svc.GetGameInfo(context.Background(), id)
		if err != nil || info == nil || info.GameID != id {
			t.Fatalf("%s: %+v err=%v", id, info, err)
		}
	}
	_, err := svc.GetGameInfo(context.Background(), "nope")
	if err == nil {
		t.Fatal("expected not found")
	}
}

func TestSubmitScore_UnknownAndFlappyValidation(t *testing.T) {
	svc := NewGameService(&mockGameRepo{}, nil, nil)
	resp, err := svc.SubmitScore(context.Background(), &SubmitScoreRequest{GameID: "nope", UserID: "u"})
	if err != nil || resp.Success {
		t.Fatalf("%+v err=%v", resp, err)
	}
	resp, err = svc.SubmitScore(context.Background(), &SubmitScoreRequest{
		GameID: "flappy", UserID: "u", Score: -1, Level: 1,
	})
	if err != nil || resp.Success {
		t.Fatalf("bad score: %+v", resp)
	}
	resp, err = svc.SubmitScore(context.Background(), &SubmitScoreRequest{
		GameID: "flappy", UserID: "u", Score: 10, Level: 99,
	})
	if err != nil || resp.Success {
		t.Fatalf("bad level: %+v", resp)
	}
}

func TestSubmitScore_FlappyNewRecord(t *testing.T) {
	repo := &mockGameRepo{highscore: 5}
	svc := NewGameService(repo, nil, nil)
	resp, err := svc.SubmitScore(context.Background(), &SubmitScoreRequest{
		UserID: "u1", GameID: "flappy", Score: 50, Level: 2, UserEmail: "a@b.c", Nickname: "n",
	})
	if err != nil || !resp.Success || !resp.Ranked || resp.LampsEarned <= 0 {
		t.Fatalf("%+v err=%v", resp, err)
	}
	if !repo.saveEnqueue {
		t.Fatal("expected SaveHighscoreAndEnqueueOutbox")
	}
}

func TestSubmitScore_FlappyNonRecord(t *testing.T) {
	repo := &mockGameRepo{highscore: 100}
	svc := NewGameService(repo, nil, nil)
	resp, err := svc.SubmitScore(context.Background(), &SubmitScoreRequest{
		UserID: "u1", GameID: "flappy", Score: 10, Level: 1,
	})
	if err != nil || !resp.Success || !repo.enqueue {
		t.Fatalf("%+v enqueue=%v err=%v", resp, repo.enqueue, err)
	}
}

func TestSubmitScore_BoostedNotRanked(t *testing.T) {
	repo := &mockGameRepo{consumeOK: true, consumeID: "b1"}
	svc := NewGameService(repo, nil, nil)
	resp, err := svc.SubmitScore(context.Background(), &SubmitScoreRequest{
		UserID: "u1", GameID: "flappy", Score: 20, Level: 1, BoostID: "b1",
	})
	if err != nil || !resp.Success || resp.Ranked || resp.LampsEarned != 0 {
		t.Fatalf("%+v err=%v", resp, err)
	}
}

func TestSubmitScore_TowersAndHanoi(t *testing.T) {
	repo := &mockGameRepo{highscore: 0}
	svc := NewGameService(repo, nil, nil)
	resp, err := svc.SubmitScore(context.Background(), &SubmitScoreRequest{
		UserID: "u", GameID: "towers", Score: 100, Level: 1,
	})
	if err != nil || !resp.Success || resp.TicketsEarned != 5 {
		t.Fatalf("towers %+v err=%v", resp, err)
	}
	resp, err = svc.SubmitScore(context.Background(), &SubmitScoreRequest{
		UserID: "u", GameID: "hanoi", Score: 950, Level: 1,
	})
	if err != nil || !resp.Success || resp.LampsEarned < 5 {
		t.Fatalf("hanoi %+v err=%v", resp, err)
	}
}

func TestSubmitScore_MoreGamesAndOutboxFallback(t *testing.T) {
	repo := &mockGameRepo{highscore: 0, err: errors.New("outbox down")}
	// err on SaveHighscoreAndEnqueueOutbox triggers fallback SaveHighscore
	svc := NewGameService(repo, nil, nil)
	for _, tc := range []struct {
		game  string
		score int
	}{
		{"twenty48", 400},
		{"gears", 100},
		{"companion", 50},
	} {
		repo.saveEnqueue, repo.saveHS, repo.enqueue = false, false, false
		repo.err = nil // happy enqueue for these
		resp, err := svc.SubmitScore(context.Background(), &SubmitScoreRequest{
			UserID: "u", GameID: tc.game, Score: tc.score, Level: 1,
		})
		if err != nil || !resp.Success {
			t.Fatalf("%s %+v err=%v", tc.game, resp, err)
		}
	}
	// force outbox failure on new record → SaveHighscore fallback (+ nil js publish skip)
	repo = &mockGameRepo{highscore: 0}
	repo.err = errors.New("outbox down")
	// SaveHighscoreAndEnqueueOutbox returns err; SaveHighscore also uses m.err — need split
	// use custom: first call fails enqueue+hs, SaveHighscore succeeds if we clear err after
	failOnce := &failEnqueueRepo{mockGameRepo: mockGameRepo{highscore: 0}}
	svc = NewGameService(failOnce, nil, nil)
	resp, err := svc.SubmitScore(context.Background(), &SubmitScoreRequest{
		UserID: "u", GameID: "flappy", Score: 30, Level: 1,
	})
	if err != nil || !resp.Success || !failOnce.savedHS {
		t.Fatalf("fallback %+v saved=%v err=%v", resp, failOnce.savedHS, err)
	}
	// non-record outbox fail → direct publish skip
	failEnq := &failEnqueueRepo{mockGameRepo: mockGameRepo{highscore: 1000}}
	svc = NewGameService(failEnq, nil, nil)
	resp, err = svc.SubmitScore(context.Background(), &SubmitScoreRequest{
		UserID: "u", GameID: "flappy", Score: 1, Level: 1,
	})
	if err != nil || !resp.Success {
		t.Fatalf("non-record fallback %+v err=%v", resp, err)
	}
}

type failEnqueueRepo struct {
	mockGameRepo
	savedHS bool
}

func (f *failEnqueueRepo) SaveHighscoreAndEnqueueOutbox(context.Context, string, string, int, int, string, []byte) error {
	return errors.New("outbox down")
}
func (f *failEnqueueRepo) SaveHighscore(context.Context, string, string, int, int) error {
	f.savedHS = true
	return nil
}
func (f *failEnqueueRepo) EnqueueOutbox(context.Context, string, []byte) error {
	return errors.New("outbox down")
}

func TestSubmitScore_ActiveBoostLookup(t *testing.T) {
	repo := &mockGameRepo{activeBoost: "ab1", consumeOK: true, consumeID: "ab1"}
	resp, err := NewGameService(repo, nil, nil).SubmitScore(context.Background(), &SubmitScoreRequest{
		UserID: "u", GameID: "flappy", Score: 10, Level: 1, // no BoostID — lookup path
	})
	if err != nil || resp.Ranked {
		t.Fatalf("%+v err=%v", resp, err)
	}
}

func TestStartBoost_BillingUnavailableAndSpendFail(t *testing.T) {
	_, err := NewGameService(&mockGameRepo{}, nil, nil).StartBoost(context.Background(), &StartBoostRequest{UserID: "u", GameID: "flappy"})
	if err == nil {
		t.Fatal("expected billing unavailable")
	}
	_, err = NewGameService(&mockGameRepo{}, nil, &stubBilling{spendErr: errors.New("insufficient balance")}).
		StartBoost(context.Background(), &StartBoostRequest{UserID: "u", GameID: "flappy"})
	if err == nil {
		t.Fatal("expected spend fail")
	}
	_, err = NewGameService(&mockGameRepo{}, nil, &stubBilling{spendOK: false}).
		StartBoost(context.Background(), &StartBoostRequest{UserID: "u", GameID: "flappy"})
	if err == nil {
		t.Fatal("expected success=false")
	}
}

func TestStartBoost_Flappy(t *testing.T) {
	repo := &mockGameRepo{}
	billing := &stubBilling{spendOK: true, spendBal: 90}
	svc := NewGameService(repo, nil, billing)
	resp, err := svc.StartBoost(context.Background(), &StartBoostRequest{UserID: "u1", GameID: "flappy"})
	if err != nil || !resp.Boosted || resp.Cost != 10 || !repo.createBoost {
		t.Fatalf("%+v err=%v", resp, err)
	}
	_, err = svc.StartBoost(context.Background(), &StartBoostRequest{UserID: "u1", GameID: "towers"})
	if err == nil {
		t.Fatal("boost only flappy")
	}
}

func TestStartBoost_ReusesActive(t *testing.T) {
	repo := &mockGameRepo{activeBoost: "existing"}
	svc := NewGameService(repo, nil, &stubBilling{spendOK: true})
	resp, err := svc.StartBoost(context.Background(), &StartBoostRequest{UserID: "u1", GameID: "flappy"})
	if err != nil || !resp.Boosted || resp.BoostID != "existing" || repo.createBoost {
		t.Fatalf("%+v create=%v err=%v", resp, repo.createBoost, err)
	}
}

// silence unused import if repository only used via interface satisfaction
var _ repository.GameRepository = (*mockGameRepo)(nil)

func TestSubmitScore_HexagonMemorySmoke(t *testing.T) {
	svc := NewGameService(&mockGameRepo{highscore: 0}, nil, nil)
	// Exercise validator branches (empty seed/moves — outcome depends on validators).
	resp, err := svc.SubmitScore(context.Background(), &SubmitScoreRequest{
		UserID: "u", GameID: "hexagon", Score: 10, Seed: "", Moves: nil,
	})
	if err != nil || resp == nil {
		t.Fatalf("%+v err=%v", resp, err)
	}
	resp, err = svc.SubmitScore(context.Background(), &SubmitScoreRequest{
		UserID: "u", GameID: "memory", Score: 10, Seed: "x", Moves: nil,
	})
	if err != nil || resp == nil {
		t.Fatalf("%+v err=%v", resp, err)
	}
}

func TestStartBoost_NilRequest(t *testing.T) {
	_, err := NewGameService(&mockGameRepo{}, nil, &stubBilling{spendOK: true}).StartBoost(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}
