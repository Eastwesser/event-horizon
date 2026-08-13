package shop

import (
	"fmt"
	"strings"
)

func (m *GetItemsRequest) Validate() error {
	return nil
}

func (m *PurchaseItemRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.UserId) == "" {
		return fmt.Errorf("user_id is required")
	}
	if strings.TrimSpace(m.ItemId) == "" {
		return fmt.Errorf("item_id is required")
	}
	return nil
}

func (m *GetInventoryRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.UserId) == "" {
		return fmt.Errorf("user_id is required")
	}
	return nil
}
