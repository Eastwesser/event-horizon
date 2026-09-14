package calendar

import (
	"errors"
	"sort"
	"sync"
	"time"
)

var (
	ErrNotFound  = errors.New("event not found")
	ErrDuplicate = errors.New("event already exists")
	dateLayout   = "2006-01-02"
)

type Event struct {
	UserID int       `json:"user_id"`
	Date   time.Time `json:"date"`
	Text   string    `json:"event"`
}

// Service stores events in memory, grouped by user and day.
type Service struct {
	mu   sync.RWMutex
	data map[int]map[string][]Event // userID -> YYYY-MM-DD -> events
}

func NewService() *Service { return &Service{data: make(map[int]map[string][]Event)} }

func normalizeDate(t time.Time) string { return t.Format(dateLayout) }

func parseDate(s string) (time.Time, error) {
	d, err := time.Parse(dateLayout, s)
	if err != nil {
		return time.Time{}, err
	}
	return d, nil
}

func (s *Service) Create(ev Event) error {
	key := normalizeDate(ev.Date)
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.data[ev.UserID] == nil {
		s.data[ev.UserID] = make(map[string][]Event)
	}
	for _, e := range s.data[ev.UserID][key] {
		if e.Text == ev.Text {
			return ErrDuplicate
		}
	}
	s.data[ev.UserID][key] = append(s.data[ev.UserID][key], ev)
	return nil
}

func (s *Service) Update(userID int, date time.Time, text, newText string) error {
	key := normalizeDate(date)
	s.mu.Lock()
	defer s.mu.Unlock()
	events := s.data[userID][key]
	for i := range events {
		if events[i].Text == text {
			events[i].Text = newText
			s.data[userID][key][i] = events[i]
			return nil
		}
	}
	return ErrNotFound
}

func (s *Service) Delete(userID int, date time.Time, text string) error {
	key := normalizeDate(date)
	s.mu.Lock()
	defer s.mu.Unlock()
	events := s.data[userID][key]
	for i := range events {
		if events[i].Text == text {
			s.data[userID][key] = append(events[:i], events[i+1:]...)
			return nil
		}
	}
	return ErrNotFound
}

func (s *Service) EventsForDay(userID int, date time.Time) []Event {
	key := normalizeDate(date)
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := append([]Event(nil), s.data[userID][key]...)
	sort.Slice(out, func(i, j int) bool { return out[i].Text < out[j].Text })
	return out
}

func (s *Service) EventsForRange(userID int, from, to time.Time) []Event {
	if to.Before(from) {
		from, to = to, from
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Event
	for d := from; !d.After(to); d = d.Add(24 * time.Hour) {
		key := normalizeDate(d)
		out = append(out, s.data[userID][key]...)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Date.Equal(out[j].Date) {
			return out[i].Text < out[j].Text
		}
		return out[i].Date.Before(out[j].Date)
	})
	return out
}

func StartOfWeek(t time.Time) time.Time {
	wd := int(t.Weekday())
	if wd == 0 {
		wd = 7
	}
	return time.Date(t.Year(), t.Month(), t.Day()-wd+1, 0, 0, 0, 0, t.Location())
}

func StartOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}
