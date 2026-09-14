package storage

import (
	"errors"
	"sort"
	"strings"
	"sync"

	"github.com/wb_technoschool/level_3/task_3/models"
)

var (
	ErrCommentNotFound = errors.New("comment not found")
)

type Storage interface {
	Create(comment *models.Comment) error
	GetByID(id string) (*models.Comment, error)
	GetByParent(parentID string) ([]*models.Comment, error)
	GetAll() ([]*models.Comment, error)
	Delete(id string) error
	Search(query string) ([]*models.Comment, error)
}

type InMemoryStorage struct {
	mu       sync.RWMutex
	comments map[string]*models.Comment
	byParent map[string][]string
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		comments: make(map[string]*models.Comment),
		byParent: make(map[string][]string),
	}
}

func (s *InMemoryStorage) Create(comment *models.Comment) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.comments[comment.ID] = comment
	if comment.ParentID != "" {
		s.byParent[comment.ParentID] = append(s.byParent[comment.ParentID], comment.ID)
	} else {
		s.byParent[""] = append(s.byParent[""], comment.ID)
	}

	return nil
}

func (s *InMemoryStorage) GetByID(id string) (*models.Comment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	comment, exists := s.comments[id]
	if !exists || comment.Deleted {
		return nil, ErrCommentNotFound
	}

	commentCopy := *comment
	return &commentCopy, nil
}

func (s *InMemoryStorage) GetByParent(parentID string) ([]*models.Comment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	childIDs := s.byParent[parentID]
	result := make([]*models.Comment, 0)

	for _, id := range childIDs {
		if comment, exists := s.comments[id]; exists && !comment.Deleted {
			commentCopy := *comment
			result = append(result, &commentCopy)
		}
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})

	return result, nil
}

func (s *InMemoryStorage) GetAll() ([]*models.Comment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*models.Comment, 0)
	for _, comment := range s.comments {
		if !comment.Deleted {
			commentCopy := *comment
			result = append(result, &commentCopy)
		}
	}

	return result, nil
}

func (s *InMemoryStorage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	comment, exists := s.comments[id]
	if !exists || comment.Deleted {
		return ErrCommentNotFound
	}

	s.deleteRecursive(id)
	return nil
}

func (s *InMemoryStorage) deleteRecursive(id string) {
	comment := s.comments[id]
	if comment == nil || comment.Deleted {
		return
	}

	comment.Deleted = true

	childIDs := s.byParent[id]
	for _, childID := range childIDs {
		s.deleteRecursive(childID)
	}
}

func (s *InMemoryStorage) Search(query string) ([]*models.Comment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	query = strings.ToLower(query)
	result := make([]*models.Comment, 0)

	for _, comment := range s.comments {
		if comment.Deleted {
			continue
		}

		if strings.Contains(strings.ToLower(comment.Content), query) ||
			strings.Contains(strings.ToLower(comment.Author), query) {
			commentCopy := *comment
			result = append(result, &commentCopy)
		}
	}

	return result, nil
}

