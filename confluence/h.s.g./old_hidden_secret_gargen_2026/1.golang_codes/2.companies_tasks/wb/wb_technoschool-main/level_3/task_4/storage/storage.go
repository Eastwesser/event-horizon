package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wb_technoschool/level_3/task_4/models"
)

type Storage interface {
	Save(image *models.Image) error
	Get(id string) (*models.Image, error)
	Delete(id string) error
	GetAll() ([]*models.Image, error)
	UpdateStatus(id string, status models.ImageStatus, errorMsg string) error
	UpdatePaths(id string, processedPath, thumbnailPath string) error
}

type InMemoryStorage struct {
	mu     sync.RWMutex
	images map[string]*models.Image
}

func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		images: make(map[string]*models.Image),
	}
}

func (s *InMemoryStorage) Save(image *models.Image) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.images[image.ID] = image
	return nil
}

func (s *InMemoryStorage) Get(id string) (*models.Image, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	image, exists := s.images[id]
	if !exists {
		return nil, fmt.Errorf("image not found")
	}
	return image, nil
}

func (s *InMemoryStorage) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.images, id)
	return nil
}

func (s *InMemoryStorage) GetAll() ([]*models.Image, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*models.Image, 0, len(s.images))
	for _, img := range s.images {
		result = append(result, img)
	}
	return result, nil
}

func (s *InMemoryStorage) UpdateStatus(id string, status models.ImageStatus, errorMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	image, exists := s.images[id]
	if !exists {
		return os.ErrNotExist
	}
	image.Status = status
	image.Error = errorMsg
	image.UpdatedAt = time.Now()
	return nil
}

func (s *InMemoryStorage) UpdatePaths(id string, processedPath, thumbnailPath string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	image, exists := s.images[id]
	if !exists {
		return os.ErrNotExist
	}
	image.ProcessedPath = processedPath
	image.ThumbnailPath = thumbnailPath
	image.UpdatedAt = time.Now()
	return nil
}

type FileStorage struct {
	basePath string
	storage  Storage
}

func NewFileStorage(basePath string, storage Storage) (*FileStorage, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}
	dirs := []string{"original", "processed", "thumbnails"}
	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(basePath, dir), 0755); err != nil {
			return nil, err
		}
	}
	return &FileStorage{
		basePath: basePath,
		storage:  storage,
	}, nil
}

func (f *FileStorage) SaveFile(id string, data []byte, subdir string, ext string) (string, error) {
	dir := filepath.Join(f.basePath, subdir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	filename := id + ext
	path := filepath.Join(dir, filename)
	return path, os.WriteFile(path, data, 0644)
}

func (f *FileStorage) GetFilePath(path string) (string, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", os.ErrNotExist
	}
	return path, nil
}

func (f *FileStorage) DeleteFile(path string) error {
	return os.Remove(path)
}

func (f *FileStorage) BasePath() string {
	return f.basePath
}
