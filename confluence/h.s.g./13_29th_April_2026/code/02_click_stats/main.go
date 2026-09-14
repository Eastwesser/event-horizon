// Click stats: unique (user, author) за календарные сутки UTC.
package main

import (
	"fmt"
	"time"
)

type ClickStore struct {
	// date(YYYY-MM-DD) → author → set(user)
	data map[string]map[string]map[string]struct{}
}

func NewClickStore() *ClickStore {
	return &ClickStore{data: make(map[string]map[string]map[string]struct{})}
}

func dayKey(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}

func (s *ClickStore) RegisterClick(userID, authorID string, at time.Time) {
	d := dayKey(at)
	byAuthor, ok := s.data[d]
	if !ok {
		byAuthor = make(map[string]map[string]struct{})
		s.data[d] = byAuthor
	}
	users, ok := byAuthor[authorID]
	if !ok {
		users = make(map[string]struct{})
		byAuthor[authorID] = users
	}
	users[userID] = struct{}{}
}

func (s *ClickStore) StatsForAuthors(authors []string, day time.Time) map[string]int {
	byAuthor := s.data[dayKey(day)]
	out := make(map[string]int, len(authors))
	if byAuthor == nil {
		return out
	}
	for _, a := range authors {
		out[a] = len(byAuthor[a])
	}
	return out
}

func main() {
	s := NewClickStore()
	day := time.Date(2026, 4, 28, 12, 0, 0, 0, time.UTC)
	s.RegisterClick("u1", "a1", day)
	s.RegisterClick("u1", "a1", day) // дубль — не считаем
	s.RegisterClick("u2", "a1", day)
	s.RegisterClick("u1", "a2", day)

	fmt.Println(s.StatsForAuthors([]string{"a1", "a2", "a3"}, day))
	// map[a1:2 a2:1 a3:0]
}
