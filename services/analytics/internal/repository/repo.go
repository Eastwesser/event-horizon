package repository

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/Eastwesser/event-horizon/services/analytics/internal/model"
	ch "github.com/Eastwesser/event-horizon/services/analytics/internal/repository/clickhouse"
)

type AnalyticsRepo struct {
	ch *ch.Client
	db string
}

func New(client *ch.Client, db string) *AnalyticsRepo {
	return &AnalyticsRepo{ch: client, db: db}
}

func (r *AnalyticsRepo) Record(ctx context.Context, userID, eventType, payload string) error {
	return r.ch.InsertEvent(ctx, userID, eventType, payload, time.Now().UTC())
}

func (r *AnalyticsRepo) DAU(ctx context.Context, days int) ([]model.DayCount, error) {
	if days <= 0 {
		days = 30
	}
	sql := fmt.Sprintf(`
		SELECT toString(event_date), uniqExact(user_id)
		FROM %s.analytics_events
		WHERE event_date >= today() - {days:Int32} AND user_id != ''
		GROUP BY event_date
		ORDER BY event_date`, r.db)
	raw, err := r.ch.QueryTSV(ctx, sql, map[string]string{"days": ch.IntParam(days)})
	if err != nil {
		return nil, err
	}
	var out []model.DayCount
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}
		n, _ := strconv.ParseInt(parts[1], 10, 64)
		out = append(out, model.DayCount{Day: parts[0], Count: n})
	}
	return out, nil
}

func (r *AnalyticsRepo) MAU(ctx context.Context, days int) (int64, error) {
	if days <= 0 {
		days = 30
	}
	sql := fmt.Sprintf(`
		SELECT uniqExact(user_id)
		FROM %s.analytics_events
		WHERE event_date >= today() - {days:Int32} AND user_id != ''`, r.db)
	raw, err := r.ch.QueryTSV(ctx, sql, map[string]string{"days": ch.IntParam(days)})
	if err != nil {
		return 0, err
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	return strconv.ParseInt(raw, 10, 64)
}

func (r *AnalyticsRepo) Retention(ctx context.Context, cohortDaysAgo, windowDays int) (*model.Retention, error) {
	if cohortDaysAgo <= 0 {
		cohortDaysAgo = 7
	}
	if windowDays <= 0 {
		windowDays = 7
	}
	cohortSQL := `SELECT toString(today() - {cohort_days_ago:Int32})`
	cohortDayRaw, err := r.ch.QueryTSV(ctx, cohortSQL, map[string]string{
		"cohort_days_ago": ch.IntParam(cohortDaysAgo),
	})
	if err != nil {
		return nil, err
	}
	cohortDay := strings.TrimSpace(cohortDayRaw)

	sizeSQL := fmt.Sprintf(`
		SELECT uniqExact(user_id) FROM %s.analytics_events
		WHERE event_date = toDate({cohort_day:String}) AND user_id != ''`, r.db)
	sizeRaw, err := r.ch.QueryTSV(ctx, sizeSQL, map[string]string{"cohort_day": cohortDay})
	if err != nil {
		return nil, err
	}
	cohortSize, _ := strconv.ParseInt(strings.TrimSpace(sizeRaw), 10, 64)

	pointsSQL := fmt.Sprintf(`
		SELECT dateDiff('day', toDate({cohort_day:String}), event_date) AS day_n, uniqExact(user_id)
		FROM %s.analytics_events
		WHERE user_id IN (
			SELECT user_id FROM %s.analytics_events
			WHERE event_date = toDate({cohort_day:String}) AND user_id != ''
		)
		AND event_date BETWEEN toDate({cohort_day:String}) AND toDate({cohort_day:String}) + {window_days:Int32}
		GROUP BY day_n
		ORDER BY day_n`, r.db, r.db)
	raw, err := r.ch.QueryTSV(ctx, pointsSQL, map[string]string{
		"cohort_day":  cohortDay,
		"window_days": ch.IntParam(windowDays),
	})
	if err != nil {
		return nil, err
	}

	byDay := map[int32]int64{}
	for _, line := range strings.Split(strings.TrimSpace(raw), "\n") {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}
		d, _ := strconv.Atoi(parts[0])
		n, _ := strconv.ParseInt(parts[1], 10, 64)
		byDay[int32(d)] = n
	}

	points := make([]model.RetentionPoint, 0, windowDays+1)
	for i := 0; i <= windowDays; i++ {
		active := byDay[int32(i)]
		rate := 0.0
		if cohortSize > 0 {
			rate = float64(active) / float64(cohortSize)
		}
		points = append(points, model.RetentionPoint{DayN: int32(i), Rate: rate})
	}
	return &model.Retention{CohortDay: cohortDay, CohortSize: cohortSize, Points: points}, nil
}
