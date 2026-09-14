package storage

import (
	"errors"
	"sync"

	"github.com/wb_technoschool/level_3/task_2/models"
)

var (
	ErrURLNotFound   = errors.New("url not found")
	ErrShortURLTaken = errors.New("short url already taken")
)

type Storage interface {
	CreateURL(url *models.URL) error
	GetURLByShort(shortURL string) (*models.URL, error)
	GetURLByLong(longURL string) (*models.URL, error)
	SaveClick(click *models.Click) error
	GetClicks(shortURL string) ([]*models.Click, error)
	GetAllURLs() ([]*models.URL, error)
}

type InMemoryStorage struct {
	mu         sync.RWMutex
	urls       map[string]*models.URL
	clicks     []*models.Click
	shortIndex map[string]string
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		urls:       make(map[string]*models.URL),
		clicks:     make([]*models.Click, 0),
		shortIndex: make(map[string]string),
	}
}

func (s *InMemoryStorage) CreateURL(url *models.URL) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.shortIndex[url.ShortURL]; exists {
		return ErrShortURLTaken
	}

	s.urls[url.ID] = url
	s.shortIndex[url.ShortURL] = url.ID
	return nil
}

func (s *InMemoryStorage) GetURLByShort(shortURL string) (*models.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	id, exists := s.shortIndex[shortURL]
	if !exists {
		return nil, ErrURLNotFound
	}

	url, exists := s.urls[id]
	if !exists {
		return nil, ErrURLNotFound
	}

	urlCopy := *url
	return &urlCopy, nil
}

func (s *InMemoryStorage) GetURLByLong(longURL string) (*models.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, url := range s.urls {
		if url.LongURL == longURL {
			urlCopy := *url
			return &urlCopy, nil
		}
	}

	return nil, ErrURLNotFound
}

func (s *InMemoryStorage) SaveClick(click *models.Click) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	clickCopy := *click
	s.clicks = append(s.clicks, &clickCopy)
	return nil
}

func (s *InMemoryStorage) GetClicks(shortURL string) ([]*models.Click, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*models.Click, 0)
	for _, click := range s.clicks {
		if click.ShortURL == shortURL {
			clickCopy := *click
			result = append(result, &clickCopy)
		}
	}

	return result, nil
}

func (s *InMemoryStorage) GetAllURLs() ([]*models.URL, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*models.URL, 0, len(s.urls))
	for _, url := range s.urls {
		urlCopy := *url
		result = append(result, &urlCopy)
	}

	return result, nil
}
