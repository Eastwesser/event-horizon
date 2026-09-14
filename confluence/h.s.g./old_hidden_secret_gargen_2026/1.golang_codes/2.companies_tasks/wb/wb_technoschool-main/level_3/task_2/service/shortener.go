package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/wb_technoschool/level_3/task_2/cache"
	"github.com/wb_technoschool/level_3/task_2/models"
	"github.com/wb_technoschool/level_3/task_2/storage"
)

type ShortenerService struct {
	storage storage.Storage
	cache   cache.Cache
	baseURL string
}

func NewShortenerService(storage storage.Storage, cache cache.Cache, baseURL string) *ShortenerService {
	return &ShortenerService{
		storage: storage,
		cache:   cache,
		baseURL: baseURL,
	}
}

func (s *ShortenerService) Shorten(ctx context.Context, req *models.ShortenRequest) (*models.URL, error) {
	if !strings.HasPrefix(req.URL, "http://") && !strings.HasPrefix(req.URL, "https://") {
		req.URL = "https://" + req.URL
	}

	var shortURL string
	if req.Custom != "" {
		shortURL = req.Custom
		existing, err := s.storage.GetURLByShort(shortURL)
		if err == nil && existing != nil {
			return nil, storage.ErrShortURLTaken
		}
	} else {
		shortURL = s.generateShortURL()
	}

	url := &models.URL{
		ID:        uuid.New().String(),
		ShortURL:  shortURL,
		LongURL:   req.URL,
		CreatedAt: time.Now(),
	}

	if err := s.storage.CreateURL(url); err != nil {
		return nil, err
	}

	s.cache.Set(ctx, "url:"+shortURL, url.LongURL, 24*time.Hour)

	return url, nil
}

func (s *ShortenerService) Redirect(ctx context.Context, shortURL string) (string, error) {
	cached, err := s.cache.Get(ctx, "url:"+shortURL)
	if err == nil && cached != "" {
		return cached, nil
	}

	url, err := s.storage.GetURLByShort(shortURL)
	if err != nil {
		return "", err
	}

	s.cache.Set(ctx, "url:"+shortURL, url.LongURL, 24*time.Hour)

	return url.LongURL, nil
}

func (s *ShortenerService) GetAnalytics(ctx context.Context, shortURL string) (*models.AnalyticsResponse, error) {
	url, err := s.storage.GetURLByShort(shortURL)
	if err != nil {
		return nil, err
	}

	clicks, err := s.storage.GetClicks(shortURL)
	if err != nil {
		return nil, err
	}

	analytics := &models.AnalyticsResponse{
		ShortURL:    shortURL,
		LongURL:     url.LongURL,
		TotalClicks: len(clicks),
		ByDay:       make(map[string]int),
		ByMonth:     make(map[string]int),
		ByUserAgent: make(map[string]int),
		Clicks:      make([]models.Click, 0),
	}

	for _, click := range clicks {
		day := click.ClickedAt.Format("2006-01-02")
		month := click.ClickedAt.Format("2006-01")
		analytics.ByDay[day]++
		analytics.ByMonth[month]++
		analytics.ByUserAgent[click.UserAgent]++
		analytics.Clicks = append(analytics.Clicks, *click)
	}

	return analytics, nil
}

func (s *ShortenerService) RecordClick(ctx context.Context, shortURL, userAgent, ip string) error {
	click := &models.Click{
		ID:        uuid.New().String(),
		ShortURL:  shortURL,
		UserAgent: userAgent,
		IP:        ip,
		ClickedAt: time.Now(),
	}

	return s.storage.SaveClick(click)
}

func (s *ShortenerService) generateShortURL() string {
	b := make([]byte, 6)
	rand.Read(b)
	short := base64.URLEncoding.EncodeToString(b)[:8]
	return strings.ReplaceAll(strings.ReplaceAll(short, "+", "-"), "/", "_")
}
