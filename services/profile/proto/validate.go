package profile

import (
	"fmt"
	"strings"
)

func (m *GetProfileRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.UserId) == "" {
		return fmt.Errorf("user_id is required")
	}
	return nil
}

func (m *UpdateProfileRequest) Validate() error {
	if m == nil || strings.TrimSpace(m.UserId) == "" {
		return fmt.Errorf("user_id is required")
	}
	return nil
}
