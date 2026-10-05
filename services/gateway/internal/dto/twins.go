package dto

import (
	"github.com/gin-gonic/gin"

	analyticsPb "github.com/Eastwesser/event-horizon/services/analytics/proto"
	authorsPb "github.com/Eastwesser/event-horizon/services/authors/proto"
	historyPb "github.com/Eastwesser/event-horizon/services/history/proto"
	leaderboardPb "github.com/Eastwesser/event-horizon/services/leaderboard/proto"
	notificationPb "github.com/Eastwesser/event-horizon/services/notification/proto"
)

// Author maps authors proto with explicit zeros (active/timestamps never omit).
func Author(a *authorsPb.Author) gin.H {
	if a == nil {
		return nil
	}
	return gin.H{
		"id":               a.GetId(),
		"user_id":          a.GetUserId(),
		"display_name":     a.GetDisplayName(),
		"bio":              a.GetBio(),
		"avatar_url":       a.GetAvatarUrl(),
		"portfolio":        a.GetPortfolio(),
		"active":           a.GetActive(),
		"created_at_unix":  a.GetCreatedAtUnix(),
		"updated_at_unix":  a.GetUpdatedAtUnix(),
		"verified_at_unix": a.GetVerifiedAtUnix(),
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

func AuthorApplication(a *authorsPb.AuthorApplication) gin.H {
	if a == nil {
		return nil
	}
	return gin.H{
		"id":               a.GetId(),
		"user_id":          a.GetUserId(),
		"status":           a.GetStatus(),
		"display_name":     a.GetDisplayName(),
		"portfolio":        a.GetPortfolio(),
		"motivation":       a.GetMotivation(),
		"contact_email":    a.GetContactEmail(),
		"created_at_unix":  a.GetCreatedAtUnix(),
		"reviewed_at_unix": a.GetReviewedAtUnix(),
		"reviewed_by":      a.GetReviewedBy(),
		"reviewer_note":    a.GetReviewerNote(),
	}
}

func AuthorApplications(list []*authorsPb.AuthorApplication) []gin.H {
	if list == nil {
		return []gin.H{}
	}
	out := make([]gin.H, 0, len(list))
	for _, a := range list {
		if h := AuthorApplication(a); h != nil {
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

func Notification(n *notificationPb.Notification) gin.H {
	if n == nil {
		return nil
	}
	return gin.H{
		"id":              n.GetId(),
		"user_id":         n.GetUserId(),
		"title":           n.GetTitle(),
		"body":            n.GetBody(),
		"link":            n.GetLink(),
		"created_at_unix": n.GetCreatedAtUnix(),
		"read_at_unix":    n.GetReadAtUnix(),
	}
}

func Notifications(list []*notificationPb.Notification) []gin.H {
	if list == nil {
		return []gin.H{}
	}
	out := make([]gin.H, 0, len(list))
	for _, n := range list {
		if h := Notification(n); h != nil {
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
	level := e.GetLevel()
	if level < 1 {
		level = 1
	}
	return gin.H{
		"rank":       e.GetRank(),
		"user_id":    e.GetUserId(),
		"user_email": e.GetUserEmail(),
		"nickname":   e.GetNickname(),
		"score":      e.GetScore(),
		"updated_at": e.GetUpdatedAt(),
		"level":      level,
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
