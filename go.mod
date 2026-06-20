module github.com/Eastwesser/event-horizon

go 1.25.7

replace (
	github.com/Eastwesser/event-horizon/services/auth/proto => ./services/auth/proto
	github.com/Eastwesser/event-horizon/services/billing/proto => ./services/billing/proto
	github.com/Eastwesser/event-horizon/services/game/proto => ./services/game/proto
	github.com/Eastwesser/event-horizon/services/leaderboard/proto => ./services/leaderboard/proto
)
