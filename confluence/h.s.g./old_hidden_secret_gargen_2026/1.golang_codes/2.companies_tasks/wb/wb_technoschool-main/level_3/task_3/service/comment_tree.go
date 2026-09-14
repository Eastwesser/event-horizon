package service

import (
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/wb_technoschool/level_3/task_3/models"
	"github.com/wb_technoschool/level_3/task_3/storage"
)

type CommentTreeService struct {
	storage storage.Storage
}

func NewCommentTreeService(storage storage.Storage) *CommentTreeService {
	return &CommentTreeService{
		storage: storage,
	}
}

func (s *CommentTreeService) CreateComment(req *models.CreateCommentRequest) (*models.Comment, error) {
	comment := &models.Comment{
		ID:        generateID(),
		ParentID:  req.ParentID,
		Author:    req.Author,
		Content:   req.Content,
		CreatedAt: timeNow(),
		UpdatedAt: timeNow(),
		Deleted:   false,
	}

	if err := s.storage.Create(comment); err != nil {
		return nil, err
	}

	return comment, nil
}

func (s *CommentTreeService) GetComments(parentID string, sortBy string, page, pageSize int) (*models.CommentsResponse, error) {
	var comments []*models.Comment
	var err error

	if parentID != "" {
		comments, err = s.storage.GetByParent(parentID)
	} else {
		all, err := s.storage.GetAll()
		if err != nil {
			return nil, err
		}
		comments = s.filterRootComments(all)
	}

	if err != nil {
		return nil, err
	}

	s.sortComments(comments, sortBy)

	tree := s.buildTree(comments)

	total := len(tree)
	totalPages := (total + pageSize - 1) / pageSize

	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}

	if start >= total {
		tree = []models.Comment{}
	} else {
		tree = tree[start:end]
	}

	return &models.CommentsResponse{
		Comments:   tree,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *CommentTreeService) DeleteComment(id string) error {
	return s.storage.Delete(id)
}

func (s *CommentTreeService) SearchComments(query string) ([]models.Comment, error) {
	comments, err := s.storage.Search(query)
	if err != nil {
		return nil, err
	}

	allComments, _ := s.storage.GetAll()
	commentMap := make(map[string]*models.Comment)
	for _, c := range allComments {
		commentMap[c.ID] = c
	}

	tree := s.buildTreeFromList(comments, commentMap)

	return tree, nil
}

func (s *CommentTreeService) filterRootComments(all []*models.Comment) []*models.Comment {
	root := make([]*models.Comment, 0)
	for _, c := range all {
		if c.ParentID == "" {
			root = append(root, c)
		}
	}
	return root
}

func (s *CommentTreeService) sortComments(comments []*models.Comment, sortBy string) {
	switch sortBy {
	case "date_desc":
		sort.Slice(comments, func(i, j int) bool {
			return comments[i].CreatedAt.After(comments[j].CreatedAt)
		})
	case "date_asc":
		sort.Slice(comments, func(i, j int) bool {
			return comments[i].CreatedAt.Before(comments[j].CreatedAt)
		})
	case "author":
		sort.Slice(comments, func(i, j int) bool {
			return comments[i].Author < comments[j].Author
		})
	default:
		sort.Slice(comments, func(i, j int) bool {
			return comments[i].CreatedAt.Before(comments[j].CreatedAt)
		})
	}
}

func (s *CommentTreeService) buildTree(comments []*models.Comment) []models.Comment {
	allComments, _ := s.storage.GetAll()
	commentMap := make(map[string]*models.Comment)
	for _, c := range allComments {
		commentMap[c.ID] = c
	}

	result := make([]models.Comment, 0)
	for _, comment := range comments {
		tree := s.buildCommentTree(comment, commentMap)
		result = append(result, tree)
	}

	return result
}

func (s *CommentTreeService) buildTreeFromList(comments []*models.Comment, commentMap map[string]*models.Comment) []models.Comment {
	result := make([]models.Comment, 0)
	added := make(map[string]bool)

	for _, comment := range comments {
		if !added[comment.ID] {
			tree := s.buildCommentTreeWithParents(comment, commentMap, added)
			result = append(result, tree)
		}
	}

	return result
}

func (s *CommentTreeService) buildCommentTree(comment *models.Comment, commentMap map[string]*models.Comment) models.Comment {
	copy := models.Comment{
		ID:        comment.ID,
		ParentID:  comment.ParentID,
		Author:    comment.Author,
		Content:   comment.Content,
		CreatedAt: comment.CreatedAt,
		UpdatedAt: comment.UpdatedAt,
		Deleted:   comment.Deleted,
		Children:  []models.Comment{},
	}

	children, _ := s.storage.GetByParent(comment.ID)
	for _, child := range children {
		childTree := s.buildCommentTree(child, commentMap)
		copy.Children = append(copy.Children, childTree)
	}

	return copy
}

func (s *CommentTreeService) buildCommentTreeWithParents(comment *models.Comment, commentMap map[string]*models.Comment, added map[string]bool) models.Comment {
	added[comment.ID] = true

	root := comment
	for root.ParentID != "" {
		if parent, exists := commentMap[root.ParentID]; exists {
			root = parent
			added[root.ID] = true
		} else {
			break
		}
	}

	return s.buildCommentTree(root, commentMap)
}

func generateID() string {
	return uuid.New().String()
}

func timeNow() time.Time {
	return time.Now()
}

