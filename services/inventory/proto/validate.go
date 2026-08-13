package inventory

import (
	"fmt"
	"strings"
)

func (m *CreateItemRequest) Validate() error {
	if m == nil {
		return fmt.Errorf("request is nil")
	}
	if strings.TrimSpace(m.AuthorId) == "" {
		return fmt.Errorf("author_id is required")
	}
	if strings.TrimSpace(m.Type) == "" {
		return fmt.Errorf("type is required")
	}
	if strings.TrimSpace(m.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if m.Price < 0 {
		return fmt.Errorf("price must be >= 0")
	}
	if m.Stock < 0 {
		return fmt.Errorf("stock must be >= 0")
	}
	if len(m.Images) > 20 {
		return fmt.Errorf("images max 20")
	}
	return nil
}

func (m *GetItemRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.Id) == "" {
		return fmt.Errorf("id is required")
	}
	return nil
}

func (m *UpdateItemRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.Id) == "" {
		return fmt.Errorf("id is required")
	}
	if m.Price < 0 {
		return fmt.Errorf("price must be >= 0")
	}
	if m.Stock < 0 {
		return fmt.Errorf("stock must be >= 0")
	}
	return nil
}

func (m *DeleteItemRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.Id) == "" {
		return fmt.Errorf("id is required")
	}
	return nil
}

func (m *SearchItemsRequest) Validate() error {
	if m == nil {
		return nil
	}
	if m.Limit < 0 || m.Limit > 100 {
		return fmt.Errorf("limit must be 0-100")
	}
	if m.Offset < 0 {
		return fmt.Errorf("offset must be >= 0")
	}
	return nil
}

func (m *GetByAuthorRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.AuthorId) == "" {
		return fmt.Errorf("author_id is required")
	}
	if m.Limit < 0 || m.Limit > 100 {
		return fmt.Errorf("limit must be 0-100")
	}
	return nil
}

func (m *GetByTypeRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.Type) == "" {
		return fmt.Errorf("type is required")
	}
	if m.Limit < 0 || m.Limit > 100 {
		return fmt.Errorf("limit must be 0-100")
	}
	return nil
}

func (m *BulkCreateItemsRequest) Validate() error {
	if m == nil || len(m.Items) == 0 {
		return fmt.Errorf("items required")
	}
	if len(m.Items) > 100 {
		return fmt.Errorf("items max 100")
	}
	for i, item := range m.Items {
		if err := item.Validate(); err != nil {
			return fmt.Errorf("items[%d]: %w", i, err)
		}
	}
	return nil
}

func (m *ReserveItemRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.Id) == "" {
		return fmt.Errorf("id is required")
	}
	if m.Quantity <= 0 {
		return fmt.Errorf("quantity must be > 0")
	}
	return nil
}

func (m *SoftDeleteItemRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.Id) == "" {
		return fmt.Errorf("id is required")
	}
	return nil
}

func (m *RestoreItemRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.Id) == "" {
		return fmt.Errorf("id is required")
	}
	return nil
}

func (m *EmptyRequest) Validate() error { return nil }
