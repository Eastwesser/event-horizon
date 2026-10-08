package service

import (
	"testing"
)

func TestEvaluateAchievements(t *testing.T) {
	codes := EvaluateAchievements(map[string]int32{
		"flappy": 120,
		"hexagon": 10,
	}, 130, "flappy", 5)

	want := map[string]bool{
		"first_play_flappy":  true,
		"first_play_hexagon": true,
		"flappy_amateur":     true,
		"flappy_pro":         true,
		"flappy_score_100":   true,
		"flappy_level_5":     true,
	}
	got := map[string]bool{}
	for _, c := range codes {
		got[c] = true
	}
	for k := range want {
		if !got[k] {
			t.Fatalf("missing %s in %v", k, codes)
		}
	}
	if got["flappy_score_500"] || got["flappy_level_10"] || got["total_score_1k"] {
		t.Fatalf("unexpected high-tier unlocks: %v", codes)
	}

	codes = EvaluateAchievements(map[string]int32{"flappy": 600}, 6000, "flappy", 10)
	got = map[string]bool{}
	for _, c := range codes {
		got[c] = true
	}
	for _, k := range []string{"flappy_score_500", "flappy_level_10", "total_score_5k", "total_score_1k"} {
		if !got[k] {
			t.Fatalf("missing %s in %v", k, codes)
		}
	}
}
