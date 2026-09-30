package dto

import (
	"github.com/gin-gonic/gin"

	analyticsPb "github.com/Eastwesser/event-horizon/services/analytics/proto"
	authorsPb "github.com/Eastwesser/event-horizon/services/authors/proto"
	historyPb "github.com/Eastwesser/event-horizon/services/history/proto"
	leaderboardPb "github.com/Eastwesser/event-horizon/services/leaderboard/proto"
)

// Author maps authors proto with explicit zeros (active/timestamps never omit).
func Author(a *authorsPb.Author) gin.H {
	if a == nil {
		return nil
	}
	return gin.H{
		"id":              a.GetId(),
		"user_id":         a.GetUserId(),
		"display_name":    a.GetDisplayName(),
		"bio":             a.GetBio(),
		"avatar_url":      a.GetAvatarUrl(),
		"active":          a.GetActive(),
		"created_at_unix": a.GetCreatedAtUnix(),
		"updated_at_unix": a.GetUpdatedAtUnix(),
	}
}

func Authors(list []*authorsPb.Author) []gin.H {
	if list == nil {
		return []gin.H{}
	}
	out := make([]gin.H, 0, len(list))
	for _, a := range list {
		if h := Author(a); h != nil {
			out = append(out, h)
		}
	}
	return out
}

func HistoryEvent(e *historyPb.HistoryEvent) gin.H {
	if e == nil {
		return nil
	}
	return gin.H{
		"id":              e.GetId(),
		"user_id":         e.GetUserId(),
		"event_type":      e.GetEventType(),
		"payload_json":    e.GetPayloadJson(),
		"created_at_unix": e.GetCreatedAtUnix(),
	}
}

func HistoryEvents(list []*historyPb.HistoryEvent) []gin.H {
	if list == nil {
		return []gin.H{}
	}
	out := make([]gin.H, 0, len(list))
	for _, e := range list {
		if h := HistoryEvent(e); h != nil {
			out = append(out, h)
		}
	}
	return out
}

func DayCount(d *analyticsPb.DayCount) gin.H {
	if d == nil {
		return nil
	}
	return gin.H{
		"day":   d.GetDay(),
		"count": d.GetCount(),
	}
}

func DayCounts(list []*analyticsPb.DayCount) []gin.H {
	if list == nil {
		return []gin.H{}
	}
	out := make([]gin.H, 0, len(list))
	for _, d := range list {
		if h := DayCount(d); h != nil {
			out = append(out, h)
		}
	}
	return out
}

func RetentionPoint(p *analyticsPb.RetentionPoint) gin.H {
	if p == nil {
		return nil
	}
	return gin.H{
		"day_n": p.GetDayN(),
		"rate":  p.GetRate(),
	}
}

func RetentionPoints(list []*analyticsPb.RetentionPoint) []gin.H {
	if list == nil {
		return []gin.H{}
	}
	out := make([]gin.H, 0, len(list))
	for _, p := range list {
		if h := RetentionPoint(p); h != nil {
			out = append(out, h)
		}
	}
	return out
}

func ScoreEntry(e *leaderboardPb.ScoreEntry) gin.H {
	if e == nil {
		return nil
	}
	return gin.H{
		"rank":       e.GetRank(),
		"user_id":    e.GetUserId(),
		"user_email": e.GetUserEmail(),
		"nickname":   e.GetNickname(),
		"score":      e.GetScore(),
		"updated_at": e.GetUpdatedAt(),
	}
}

func ScoreEntries(list []*leaderboardPb.ScoreEntry) []gin.H {
	if list == nil {
		return []gin.H{}
	}
	out := make([]gin.H, 0, len(list))
	for _, e := range list {
		if h := ScoreEntry(e); h != nil {
			out = append(out, h)
		}
	}
	return out
}
